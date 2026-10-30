package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jakaria9001/tablebite/internal/auth"
	"github.com/jakaria9001/tablebite/internal/notification"
	"github.com/jakaria9001/tablebite/internal/repository"
)

const (
	defaultOTPExpiry      = 5 * time.Minute
	defaultResendCooldown = 30 * time.Second
	defaultMaxAttempts    = 5
	defaultSessionExpiry  = 24 * time.Hour
	otpDigits             = 6
)

var (
	ErrInvalidPhone        = errors.New("invalid Indian mobile number")
	ErrNoOTPChallenge      = errors.New("OTP challenge not found")
	ErrInvalidOTP          = errors.New("invalid OTP")
	ErrOTPExpired          = errors.New("OTP has expired")
	ErrOTPConsumed         = errors.New("OTP has already been used")
	ErrOTPAttemptsExceeded = errors.New("maximum OTP attempts exceeded")
	ErrOTPResendCooldown   = errors.New("OTP resend cooldown active")
	sixDigitPattern        = regexp.MustCompile(`^[0-9]{6}$`)
	phoneDigitsPattern     = regexp.MustCompile(`^[0-9]+$`)
)

type OTPConfig struct {
	Provider       string
	Expiry         time.Duration
	ResendCooldown time.Duration
	MaxAttempts    int
	Pepper         string
	SessionExpiry  time.Duration
	Now            func() time.Time
}

func LoadOTPConfig() (OTPConfig, error) {
	config := OTPConfig{
		Provider:       getenv("OTP_PROVIDER", "mock"),
		Expiry:         defaultOTPExpiry,
		ResendCooldown: defaultResendCooldown,
		MaxAttempts:    defaultMaxAttempts,
		Pepper:         os.Getenv("OTP_PEPPER"),
		SessionExpiry:  defaultSessionExpiry,
		Now:            time.Now,
	}

	var err error
	if config.Expiry, err = durationFromSeconds("OTP_EXPIRY_SECONDS", config.Expiry); err != nil {
		return OTPConfig{}, err
	}
	if config.ResendCooldown, err = durationFromSeconds("OTP_RESEND_COOLDOWN_SECONDS", config.ResendCooldown); err != nil {
		return OTPConfig{}, err
	}
	if config.MaxAttempts, err = positiveIntFromEnv("OTP_MAX_ATTEMPTS", config.MaxAttempts); err != nil {
		return OTPConfig{}, err
	}
	if config.SessionExpiry, err = durationFromHours("OTP_SESSION_EXPIRY_HOURS", config.SessionExpiry); err != nil {
		return OTPConfig{}, err
	}
	if strings.TrimSpace(config.Pepper) == "" {
		return OTPConfig{}, errors.New("OTP_PEPPER is required")
	}
	return config, nil
}

type OTPStore interface {
	CreateChallenge(ctx context.Context, phone string, hash []byte, expiresAt time.Time, maxAttempts int, resendCooldown time.Duration, now time.Time) (repository.OTPChallenge, error)
	FindLatestChallenge(ctx context.Context, phone string) (repository.OTPChallenge, bool, error)
	IncrementAttempt(ctx context.Context, id int64) (int, error)
	ConsumeChallenge(ctx context.Context, id int64, verifiedAt *time.Time) (bool, error)
}

type OTPVerificationStore interface {
	OTPStore
	VerifyChallengeAndCreateSession(ctx context.Context, phone string, otpHash []byte, now time.Time, sessionHash []byte, sessionExpires time.Time) (repository.VerifiedCustomerSession, error)
}

type OTPService struct {
	store    OTPStore
	provider notification.OTPProvider
	config   OTPConfig
}

type OTPVerificationResult struct {
	Customer       repository.Customer
	SessionToken   string
	SessionExpires time.Time
}

func NewOTPService(store OTPStore, provider notification.OTPProvider, config OTPConfig) (*OTPService, error) {
	if store == nil {
		return nil, errors.New("OTP store is required")
	}
	if provider == nil {
		return nil, errors.New("OTP provider is required")
	}
	if strings.TrimSpace(config.Pepper) == "" {
		return nil, errors.New("OTP pepper is required")
	}
	if config.Expiry <= 0 || config.ResendCooldown < 0 || config.MaxAttempts <= 0 {
		return nil, errors.New("invalid OTP configuration")
	}
	if config.Now == nil {
		config.Now = time.Now
	}
	return &OTPService{store: store, provider: provider, config: config}, nil
}

type OTPRequestResult struct {
	ChallengeID int64
	ExpiresAt   time.Time
	RetryAfter  time.Duration
}

func (s *OTPService) RequestOTP(ctx context.Context, phone string) (OTPRequestResult, error) {
	phone, err := NormalizePhone(phone)
	if err != nil {
		return OTPRequestResult{}, err
	}

	now := s.config.Now()
	previous, found, err := s.store.FindLatestChallenge(ctx, phone)
	if err != nil {
		return OTPRequestResult{}, err
	}
	if found {
		retryAfter := previous.CreatedAt.Add(s.config.ResendCooldown).Sub(now)
		if retryAfter > 0 {
			return OTPRequestResult{RetryAfter: retryAfter}, ErrOTPResendCooldown
		}
	}

	code, err := generateOTP()
	if err != nil {
		return OTPRequestResult{}, fmt.Errorf("generate OTP: %w", err)
	}

	expiresAt := now.Add(s.config.Expiry)
	hash := hashOTP(s.config.Pepper, phone, code)
	challenge, err := s.store.CreateChallenge(ctx, phone, hash, expiresAt, s.config.MaxAttempts, s.config.ResendCooldown, now)
	if errors.Is(err, repository.ErrOTPChallengeCooldown) {
		return OTPRequestResult{RetryAfter: s.config.ResendCooldown}, ErrOTPResendCooldown
	}
	if err != nil {
		return OTPRequestResult{}, err
	}
	if err := s.provider.SendOTP(ctx, phone, code); err != nil {
		_, consumeErr := s.store.ConsumeChallenge(ctx, challenge.ID, nil)
		if consumeErr != nil {
			return OTPRequestResult{}, fmt.Errorf("send OTP: %w; invalidate failed challenge: %v", err, consumeErr)
		}
		return OTPRequestResult{}, fmt.Errorf("send OTP: %w", err)
	}
	return OTPRequestResult{
		ChallengeID: challenge.ID,
		ExpiresAt:   challenge.ExpiresAt,
	}, nil
}

func (s *OTPService) VerifyOTP(ctx context.Context, phone, code string) error {
	phone, err := NormalizePhone(phone)
	if err != nil {
		return err
	}
	if !sixDigitPattern.MatchString(strings.TrimSpace(code)) {
		return ErrInvalidOTP
	}

	challenge, found, err := s.store.FindLatestChallenge(ctx, phone)
	if err != nil {
		return err
	}
	if !found {
		return ErrNoOTPChallenge
	}
	now := s.config.Now()
	if challenge.ConsumedAt != nil {
		if challenge.VerifiedAt != nil {
			return ErrOTPConsumed
		}
		if !challenge.ExpiresAt.After(now) {
			return ErrOTPExpired
		}
		return ErrOTPAttemptsExceeded
	}
	if !challenge.ExpiresAt.After(now) {
		_, consumeErr := s.store.ConsumeChallenge(ctx, challenge.ID, nil)
		if consumeErr != nil {
			return consumeErr
		}
		return ErrOTPExpired
	}

	attempts, err := s.store.IncrementAttempt(ctx, challenge.ID)
	if errors.Is(err, repository.ErrOTPChallengeUnavailable) {
		return ErrOTPAttemptsExceeded
	}
	if err != nil {
		return err
	}
	if attempts > challenge.MaxAttempts {
		return ErrOTPAttemptsExceeded
	}

	expected := hashOTP(s.config.Pepper, phone, strings.TrimSpace(code))
	if !hmac.Equal(expected, challenge.OTPHash) {
		if attempts >= challenge.MaxAttempts {
			_, err := s.store.ConsumeChallenge(ctx, challenge.ID, nil)
			if err != nil {
				return err
			}
			return ErrOTPAttemptsExceeded
		}
		return ErrInvalidOTP
	}

	verifiedAt := now
	consumed, err := s.store.ConsumeChallenge(ctx, challenge.ID, &verifiedAt)
	if err != nil {
		return err
	}
	if !consumed {
		return ErrOTPConsumed
	}
	return nil
}

func (s *OTPService) VerifyOTPAndCreateSession(ctx context.Context, phone, code string) (OTPVerificationResult, error) {
	phone, err := NormalizePhone(phone)
	if err != nil {
		return OTPVerificationResult{}, err
	}
	code = strings.TrimSpace(code)
	if !sixDigitPattern.MatchString(code) {
		return OTPVerificationResult{}, ErrInvalidOTP
	}

	store, ok := s.store.(OTPVerificationStore)
	if !ok {
		return OTPVerificationResult{}, errors.New("OTP store does not support customer sessions")
	}
	now := s.config.Now()
	sessionToken, err := auth.GenerateSessionToken()
	if err != nil {
		return OTPVerificationResult{}, fmt.Errorf("generate customer session token: %w", err)
	}
	sessionHash := sha256.Sum256([]byte(sessionToken))
	verification, err := store.VerifyChallengeAndCreateSession(
		ctx,
		phone,
		hashOTP(s.config.Pepper, phone, code),
		now,
		sessionHash[:],
		now.Add(s.config.SessionExpiry),
	)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrOTPChallengeNotFound):
			return OTPVerificationResult{}, ErrNoOTPChallenge
		case errors.Is(err, repository.ErrOTPChallengeExpired):
			return OTPVerificationResult{}, ErrOTPExpired
		case errors.Is(err, repository.ErrOTPChallengeConsumed):
			return OTPVerificationResult{}, ErrOTPConsumed
		case errors.Is(err, repository.ErrOTPChallengeAttemptsExceeded):
			return OTPVerificationResult{}, ErrOTPAttemptsExceeded
		case errors.Is(err, repository.ErrOTPCodeInvalid):
			return OTPVerificationResult{}, ErrInvalidOTP
		default:
			return OTPVerificationResult{}, err
		}
	}
	return OTPVerificationResult{
		Customer:       verification.Customer,
		SessionToken:   sessionToken,
		SessionExpires: verification.SessionExpires,
	}, nil
}

func NormalizePhone(phone string) (string, error) {
	value := strings.TrimSpace(phone)
	value = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(value)
	switch {
	case strings.HasPrefix(value, "+91"):
		value = value[1:]
	case strings.HasPrefix(value, "91") && len(value) == 12:
	case len(value) == 10:
		value = "91" + value
	default:
		return "", ErrInvalidPhone
	}
	if len(value) != 12 || value[:2] != "91" || value[2] < '6' || value[2] > '9' || !phoneDigitsPattern.MatchString(value) {
		return "", ErrInvalidPhone
	}
	return "+" + value, nil
}

func generateOTP() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", otpDigits, value.Int64()), nil
}

func hashOTP(pepper, phone, code string) []byte {
	mac := hmac.New(sha256.New, []byte(pepper))
	_, _ = mac.Write([]byte(phone))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(code))
	return mac.Sum(nil)
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func durationFromSeconds(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return time.Duration(seconds) * time.Second, nil
}

func durationFromHours(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	hours, err := strconv.Atoi(value)
	if err != nil || hours <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return time.Duration(hours) * time.Hour, nil
}

func positiveIntFromEnv(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return parsed, nil
}
