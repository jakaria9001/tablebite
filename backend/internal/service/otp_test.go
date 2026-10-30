package service

import (
	"context"
	"crypto/hmac"
	"errors"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/jakaria9001/tablebite/internal/repository"
)

type fakeOTPStore struct {
	now        time.Time
	nextID     int64
	challenges []repository.OTPChallenge
	customers  int
	mu         sync.Mutex
}

func (s *fakeOTPStore) CreateChallenge(_ context.Context, phone string, hash []byte, expiresAt time.Time, maxAttempts int, resendCooldown time.Duration, now time.Time) (repository.OTPChallenge, error) {
	if len(s.challenges) > 0 {
		previous := s.challenges[len(s.challenges)-1]
		if previous.CreatedAt.Add(resendCooldown).After(now) {
			return repository.OTPChallenge{}, repository.ErrOTPChallengeCooldown
		}
	}
	for index := range s.challenges {
		if s.challenges[index].PhoneNumber == phone && s.challenges[index].ConsumedAt == nil {
			consumedAt := s.now
			s.challenges[index].ConsumedAt = &consumedAt
		}
	}
	s.nextID++
	challenge := repository.OTPChallenge{
		ID:          s.nextID,
		PhoneNumber: phone,
		OTPHash:     hash,
		ExpiresAt:   expiresAt,
		MaxAttempts: maxAttempts,
		CreatedAt:   s.now,
	}
	s.challenges = append(s.challenges, challenge)
	return challenge, nil
}

func (s *fakeOTPStore) FindLatestChallenge(_ context.Context, phone string) (repository.OTPChallenge, bool, error) {
	for index := len(s.challenges) - 1; index >= 0; index-- {
		if s.challenges[index].PhoneNumber == phone {
			return s.challenges[index], true, nil
		}
	}
	return repository.OTPChallenge{}, false, nil
}

func (s *fakeOTPStore) IncrementAttempt(_ context.Context, id int64) (int, error) {
	for index := range s.challenges {
		challenge := &s.challenges[index]
		if challenge.ID == id && challenge.ConsumedAt == nil && challenge.AttemptCount < challenge.MaxAttempts {
			challenge.AttemptCount++
			return challenge.AttemptCount, nil
		}
	}
	return 0, repository.ErrOTPChallengeUnavailable
}

func (s *fakeOTPStore) ConsumeChallenge(_ context.Context, id int64, verifiedAt *time.Time) (bool, error) {
	for index := range s.challenges {
		challenge := &s.challenges[index]
		if challenge.ID == id && challenge.ConsumedAt == nil {
			consumedAt := s.now
			challenge.ConsumedAt = &consumedAt
			challenge.VerifiedAt = verifiedAt
			return true, nil
		}
	}
	return false, nil
}

func (s *fakeOTPStore) VerifyChallengeAndCreateSession(_ context.Context, phone string, otpHash []byte, now time.Time, _ []byte, sessionExpires time.Time) (repository.VerifiedCustomerSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index := len(s.challenges) - 1; index >= 0; index-- {
		challenge := &s.challenges[index]
		if challenge.PhoneNumber != phone {
			continue
		}
		if challenge.ConsumedAt != nil {
			if challenge.VerifiedAt != nil {
				return repository.VerifiedCustomerSession{}, repository.ErrOTPChallengeConsumed
			}
			return repository.VerifiedCustomerSession{}, repository.ErrOTPChallengeAttemptsExceeded
		}
		if !challenge.ExpiresAt.After(now) {
			consumedAt := now
			challenge.ConsumedAt = &consumedAt
			return repository.VerifiedCustomerSession{}, repository.ErrOTPChallengeExpired
		}
		if challenge.AttemptCount >= challenge.MaxAttempts {
			consumedAt := now
			challenge.ConsumedAt = &consumedAt
			return repository.VerifiedCustomerSession{}, repository.ErrOTPChallengeAttemptsExceeded
		}
		challenge.AttemptCount++
		if !hmac.Equal(challenge.OTPHash, otpHash) {
			if challenge.AttemptCount >= challenge.MaxAttempts {
				consumedAt := now
				challenge.ConsumedAt = &consumedAt
				return repository.VerifiedCustomerSession{}, repository.ErrOTPChallengeAttemptsExceeded
			}
			return repository.VerifiedCustomerSession{}, repository.ErrOTPCodeInvalid
		}
		verifiedAt := now
		challenge.VerifiedAt = &verifiedAt
		challenge.ConsumedAt = &verifiedAt
		s.customers++
		return repository.VerifiedCustomerSession{
			Customer: repository.Customer{
				ID:            int64(s.customers),
				PhoneNumber:   phone,
				PhoneVerified: true,
				IsActive:      true,
				LastLoginAt:   now,
			},
			SessionExpires: sessionExpires,
		}, nil
	}
	return repository.VerifiedCustomerSession{}, repository.ErrOTPChallengeNotFound
}

type fakeOTPProvider struct {
	code string
	err  error
}

func (p *fakeOTPProvider) SendOTP(_ context.Context, _ string, code string) error {
	p.code = code
	return p.err
}

func newTestOTPService(t *testing.T, store *fakeOTPStore, provider *fakeOTPProvider, now *time.Time) *OTPService {
	t.Helper()
	service, err := NewOTPService(store, provider, OTPConfig{
		Expiry:         5 * time.Minute,
		ResendCooldown: 30 * time.Second,
		MaxAttempts:    5,
		Pepper:         "test-pepper",
		SessionExpiry:  24 * time.Hour,
		Now:            func() time.Time { return *now },
	})
	if err != nil {
		t.Fatalf("create OTP service: %v", err)
	}
	return service
}

func TestVerifyOTPAndCreateSession(t *testing.T) {
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	store := &fakeOTPStore{now: now}
	provider := &fakeOTPProvider{}
	otp := newTestOTPService(t, store, provider, &now)

	if _, err := otp.RequestOTP(context.Background(), "9876543210"); err != nil {
		t.Fatalf("request OTP: %v", err)
	}
	result, err := otp.VerifyOTPAndCreateSession(context.Background(), "9876543210", provider.code)
	if err != nil {
		t.Fatalf("verify OTP: %v", err)
	}
	if result.Customer.PhoneNumber != "+919876543210" || result.SessionToken == "" {
		t.Fatalf("verification result = %#v", result)
	}
}

func TestVerifyOTPAndCreateSessionRejectsExpiredOTP(t *testing.T) {
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	store := &fakeOTPStore{now: now}
	provider := &fakeOTPProvider{}
	otp := newTestOTPService(t, store, provider, &now)
	if _, err := otp.RequestOTP(context.Background(), "9876543210"); err != nil {
		t.Fatalf("request OTP: %v", err)
	}
	now = now.Add(5 * time.Minute)
	if _, err := otp.VerifyOTPAndCreateSession(context.Background(), "9876543210", provider.code); !errors.Is(err, ErrOTPExpired) {
		t.Fatalf("expired OTP = %v, want %v", err, ErrOTPExpired)
	}
}

func TestVerifyOTPAndCreateSessionRejectsWrongOTP(t *testing.T) {
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	store := &fakeOTPStore{now: now}
	provider := &fakeOTPProvider{}
	otp := newTestOTPService(t, store, provider, &now)
	if _, err := otp.RequestOTP(context.Background(), "9876543210"); err != nil {
		t.Fatalf("request OTP: %v", err)
	}
	if _, err := otp.VerifyOTPAndCreateSession(context.Background(), "9876543210", "000000"); !errors.Is(err, ErrInvalidOTP) {
		t.Fatalf("wrong OTP = %v, want %v", err, ErrInvalidOTP)
	}
}

func TestVerifyOTPAndCreateSessionRejectsReplay(t *testing.T) {
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	store := &fakeOTPStore{now: now}
	provider := &fakeOTPProvider{}
	otp := newTestOTPService(t, store, provider, &now)
	if _, err := otp.RequestOTP(context.Background(), "9876543210"); err != nil {
		t.Fatalf("request OTP: %v", err)
	}
	if _, err := otp.VerifyOTPAndCreateSession(context.Background(), "9876543210", provider.code); err != nil {
		t.Fatalf("first verification: %v", err)
	}
	if _, err := otp.VerifyOTPAndCreateSession(context.Background(), "9876543210", provider.code); !errors.Is(err, ErrOTPConsumed) {
		t.Fatalf("replayed OTP = %v, want %v", err, ErrOTPConsumed)
	}
}

func TestVerifyOTPAndCreateSessionEnforcesAttempts(t *testing.T) {
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	store := &fakeOTPStore{now: now}
	provider := &fakeOTPProvider{}
	otp := newTestOTPService(t, store, provider, &now)
	if _, err := otp.RequestOTP(context.Background(), "9876543210"); err != nil {
		t.Fatalf("request OTP: %v", err)
	}
	for attempt := 1; attempt <= 5; attempt++ {
		_, err := otp.VerifyOTPAndCreateSession(context.Background(), "9876543210", "000000")
		if attempt < 5 && !errors.Is(err, ErrInvalidOTP) {
			t.Fatalf("attempt %d = %v, want invalid OTP", attempt, err)
		}
		if attempt == 5 && !errors.Is(err, ErrOTPAttemptsExceeded) {
			t.Fatalf("attempt %d = %v, want attempts exceeded", attempt, err)
		}
	}
}

func TestVerifyOTPAndCreateSessionAllowsOnlyOneConcurrentVerification(t *testing.T) {
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	store := &fakeOTPStore{now: now}
	provider := &fakeOTPProvider{}
	otp := newTestOTPService(t, store, provider, &now)
	if _, err := otp.RequestOTP(context.Background(), "9876543210"); err != nil {
		t.Fatalf("request OTP: %v", err)
	}

	results := make(chan error, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := otp.VerifyOTPAndCreateSession(context.Background(), "9876543210", provider.code)
			results <- err
		}()
	}
	wait.Wait()
	close(results)

	successes := 0
	for err := range results {
		if err == nil {
			successes++
			continue
		}
		if !errors.Is(err, ErrOTPConsumed) {
			t.Fatalf("concurrent verification error = %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("successful concurrent verifications = %d, want 1", successes)
	}
}
func TestGenerateOTPProducesSixDigits(t *testing.T) {
	pattern := regexp.MustCompile(`^[0-9]{6}$`)
	for range 100 {
		code, err := generateOTP()
		if err != nil {
			t.Fatalf("generate OTP: %v", err)
		}
		if !pattern.MatchString(code) {
			t.Fatalf("OTP %q is not six digits", code)
		}
	}
}

func TestOTPExpiry(t *testing.T) {
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	store := &fakeOTPStore{now: now}
	provider := &fakeOTPProvider{}
	otp := newTestOTPService(t, store, provider, &now)

	if _, err := otp.RequestOTP(context.Background(), "9876543210"); err != nil {
		t.Fatalf("request OTP: %v", err)
	}
	now = now.Add(5 * time.Minute)
	if err := otp.VerifyOTP(context.Background(), "9876543210", provider.code); !errors.Is(err, ErrOTPExpired) {
		t.Fatalf("verify expired OTP = %v, want %v", err, ErrOTPExpired)
	}
}

func TestOTPInvalidAndValidVerification(t *testing.T) {
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	store := &fakeOTPStore{now: now}
	provider := &fakeOTPProvider{}
	otp := newTestOTPService(t, store, provider, &now)

	if _, err := otp.RequestOTP(context.Background(), "9876543210"); err != nil {
		t.Fatalf("request OTP: %v", err)
	}
	if err := otp.VerifyOTP(context.Background(), "9876543210", "000000"); !errors.Is(err, ErrInvalidOTP) {
		t.Fatalf("invalid OTP = %v, want %v", err, ErrInvalidOTP)
	}
	if err := otp.VerifyOTP(context.Background(), "9876543210", provider.code); err != nil {
		t.Fatalf("valid OTP: %v", err)
	}
}

func TestOTPCannotBeReused(t *testing.T) {
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	store := &fakeOTPStore{now: now}
	provider := &fakeOTPProvider{}
	otp := newTestOTPService(t, store, provider, &now)

	if _, err := otp.RequestOTP(context.Background(), "9876543210"); err != nil {
		t.Fatalf("request OTP: %v", err)
	}
	if err := otp.VerifyOTP(context.Background(), "9876543210", provider.code); err != nil {
		t.Fatalf("verify OTP: %v", err)
	}
	if err := otp.VerifyOTP(context.Background(), "9876543210", provider.code); !errors.Is(err, ErrOTPConsumed) {
		t.Fatalf("reused OTP = %v, want %v", err, ErrOTPConsumed)
	}
}

func TestOTPMaximumAttempts(t *testing.T) {
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	store := &fakeOTPStore{now: now}
	provider := &fakeOTPProvider{}
	otp := newTestOTPService(t, store, provider, &now)

	if _, err := otp.RequestOTP(context.Background(), "9876543210"); err != nil {
		t.Fatalf("request OTP: %v", err)
	}
	for attempt := 1; attempt <= 5; attempt++ {
		err := otp.VerifyOTP(context.Background(), "9876543210", "000000")
		if attempt < 5 && !errors.Is(err, ErrInvalidOTP) {
			t.Fatalf("attempt %d = %v, want invalid OTP", attempt, err)
		}
		if attempt == 5 && !errors.Is(err, ErrOTPAttemptsExceeded) {
			t.Fatalf("attempt %d = %v, want attempts exceeded", attempt, err)
		}
	}
	if err := otp.VerifyOTP(context.Background(), "9876543210", provider.code); !errors.Is(err, ErrOTPAttemptsExceeded) {
		t.Fatalf("locked OTP = %v, want %v", err, ErrOTPAttemptsExceeded)
	}
}

func TestOTPResendCooldown(t *testing.T) {
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	store := &fakeOTPStore{now: now}
	provider := &fakeOTPProvider{}
	otp := newTestOTPService(t, store, provider, &now)

	if _, err := otp.RequestOTP(context.Background(), "9876543210"); err != nil {
		t.Fatalf("request OTP: %v", err)
	}
	if _, err := otp.RequestOTP(context.Background(), "9876543210"); !errors.Is(err, ErrOTPResendCooldown) {
		t.Fatalf("resend = %v, want %v", err, ErrOTPResendCooldown)
	}
	now = now.Add(30 * time.Second)
	if _, err := otp.RequestOTP(context.Background(), "9876543210"); err != nil {
		t.Fatalf("request after cooldown: %v", err)
	}
}

func TestOTPProviderFailureDoesNotCreateChallenge(t *testing.T) {
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	store := &fakeOTPStore{now: now}
	providerError := errors.New("SMS provider unavailable")
	provider := &fakeOTPProvider{err: providerError}
	otp := newTestOTPService(t, store, provider, &now)

	_, err := otp.RequestOTP(context.Background(), "9876543210")
	if !errors.Is(err, providerError) {
		t.Fatalf("provider failure = %v, want wrapped provider error", err)
	}
	if len(store.challenges) != 1 || store.challenges[0].ConsumedAt == nil {
		t.Fatalf("provider failure should consume its challenge: %#v", store.challenges)
	}
}
