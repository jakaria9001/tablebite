package repository

import (
	"context"
	"crypto/hmac"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrOTPChallengeUnavailable = errors.New("OTP challenge unavailable")
var ErrOTPChallengeCooldown = errors.New("OTP challenge cooldown active")
var ErrOTPChallengeNotFound = errors.New("OTP challenge not found")
var ErrOTPChallengeExpired = errors.New("OTP challenge expired")
var ErrOTPChallengeConsumed = errors.New("OTP challenge consumed")
var ErrOTPChallengeAttemptsExceeded = errors.New("OTP challenge attempts exceeded")
var ErrOTPCodeInvalid = errors.New("OTP code invalid")

type OTPChallenge struct {
	ID           int64
	PhoneNumber  string
	OTPHash      []byte
	ExpiresAt    time.Time
	AttemptCount int
	MaxAttempts  int
	VerifiedAt   *time.Time
	ConsumedAt   *time.Time
	CreatedAt    time.Time
}

type OTPStore struct {
	db *pgxpool.Pool
}

type Customer struct {
	ID            int64
	PhoneNumber   string
	PhoneVerified bool
	Name          *string
	IsActive      bool
	LastLoginAt   time.Time
}

type VerifiedCustomerSession struct {
	Customer       Customer
	SessionExpires time.Time
}

func NewOTPStore(db *pgxpool.Pool) *OTPStore {
	return &OTPStore{db: db}
}

func (s *OTPStore) CreateChallenge(ctx context.Context, phone string, hash []byte, expiresAt time.Time, maxAttempts int, resendCooldown time.Duration, now time.Time) (OTPChallenge, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return OTPChallenge{}, fmt.Errorf("begin OTP challenge transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var previousCreatedAt time.Time
	err = tx.QueryRow(ctx, `
		SELECT created_at
		FROM otp_challenges
		WHERE phone_number = $1
		ORDER BY created_at DESC, id DESC
		LIMIT 1
		FOR UPDATE`, phone).Scan(&previousCreatedAt)
	if err == nil && previousCreatedAt.Add(resendCooldown).After(now) {
		return OTPChallenge{}, ErrOTPChallengeCooldown
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return OTPChallenge{}, fmt.Errorf("check OTP challenge cooldown: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE otp_challenges
		SET consumed_at = NOW()
		WHERE phone_number = $1 AND consumed_at IS NULL`, phone); err != nil {
		return OTPChallenge{}, fmt.Errorf("invalidate previous OTP challenges: %w", err)
	}

	var challenge OTPChallenge
	err = tx.QueryRow(ctx, `
		INSERT INTO otp_challenges (phone_number, otp_hash, expires_at, max_attempts)
		VALUES ($1, $2, $3, $4)
		RETURNING id, phone_number, otp_hash, expires_at, attempt_count, max_attempts,
			verified_at, consumed_at, created_at`,
		phone, hash, expiresAt, maxAttempts,
	).Scan(
		&challenge.ID,
		&challenge.PhoneNumber,
		&challenge.OTPHash,
		&challenge.ExpiresAt,
		&challenge.AttemptCount,
		&challenge.MaxAttempts,
		&challenge.VerifiedAt,
		&challenge.ConsumedAt,
		&challenge.CreatedAt,
	)
	if err != nil {
		return OTPChallenge{}, fmt.Errorf("create OTP challenge: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return OTPChallenge{}, fmt.Errorf("commit OTP challenge: %w", err)
	}
	return challenge, nil
}

func (s *OTPStore) FindLatestChallenge(ctx context.Context, phone string) (OTPChallenge, bool, error) {
	var challenge OTPChallenge
	err := s.db.QueryRow(ctx, `
		SELECT id, phone_number, otp_hash, expires_at, attempt_count, max_attempts,
			verified_at, consumed_at, created_at
		FROM otp_challenges
		WHERE phone_number = $1
		ORDER BY created_at DESC, id DESC
		LIMIT 1`, phone,
	).Scan(
		&challenge.ID,
		&challenge.PhoneNumber,
		&challenge.OTPHash,
		&challenge.ExpiresAt,
		&challenge.AttemptCount,
		&challenge.MaxAttempts,
		&challenge.VerifiedAt,
		&challenge.ConsumedAt,
		&challenge.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return OTPChallenge{}, false, nil
	}
	if err != nil {
		return OTPChallenge{}, false, fmt.Errorf("find latest OTP challenge: %w", err)
	}
	return challenge, true, nil
}

func (s *OTPStore) IncrementAttempt(ctx context.Context, id int64) (int, error) {
	var attempts int
	err := s.db.QueryRow(ctx, `
		UPDATE otp_challenges
		SET attempt_count = attempt_count + 1
		WHERE id = $1 AND consumed_at IS NULL AND attempt_count < max_attempts
		RETURNING attempt_count`, id).Scan(&attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrOTPChallengeUnavailable
	}
	if err != nil {
		return 0, fmt.Errorf("increment OTP attempts: %w", err)
	}
	return attempts, nil
}

func (s *OTPStore) ConsumeChallenge(ctx context.Context, id int64, verifiedAt *time.Time) (bool, error) {
	tag, err := s.db.Exec(ctx, `
		UPDATE otp_challenges
		SET consumed_at = NOW(), verified_at = $2
		WHERE id = $1 AND consumed_at IS NULL`, id, verifiedAt)
	if err != nil {
		return false, fmt.Errorf("consume OTP challenge: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

func (s *OTPStore) VerifyChallengeAndCreateSession(
	ctx context.Context,
	phone string,
	otpHash []byte,
	now time.Time,
	sessionHash []byte,
	sessionExpires time.Time,
) (VerifiedCustomerSession, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return VerifiedCustomerSession{}, fmt.Errorf("begin OTP verification transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var challenge OTPChallenge
	err = tx.QueryRow(ctx, `
		SELECT id, phone_number, otp_hash, expires_at, attempt_count, max_attempts,
			verified_at, consumed_at, created_at
		FROM otp_challenges
		WHERE phone_number = $1
		ORDER BY created_at DESC, id DESC
		LIMIT 1
		FOR UPDATE`, phone,
	).Scan(
		&challenge.ID,
		&challenge.PhoneNumber,
		&challenge.OTPHash,
		&challenge.ExpiresAt,
		&challenge.AttemptCount,
		&challenge.MaxAttempts,
		&challenge.VerifiedAt,
		&challenge.ConsumedAt,
		&challenge.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return VerifiedCustomerSession{}, ErrOTPChallengeNotFound
	}
	if err != nil {
		return VerifiedCustomerSession{}, fmt.Errorf("load OTP challenge for verification: %w", err)
	}
	if challenge.ConsumedAt != nil {
		if challenge.VerifiedAt != nil {
			return VerifiedCustomerSession{}, ErrOTPChallengeConsumed
		}
		return VerifiedCustomerSession{}, ErrOTPChallengeAttemptsExceeded
	}
	if !challenge.ExpiresAt.After(now) {
		if _, err := tx.Exec(ctx, `UPDATE otp_challenges SET consumed_at = $2 WHERE id = $1`, challenge.ID, now); err != nil {
			return VerifiedCustomerSession{}, fmt.Errorf("consume expired OTP challenge: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return VerifiedCustomerSession{}, fmt.Errorf("commit expired OTP challenge: %w", err)
		}
		return VerifiedCustomerSession{}, ErrOTPChallengeExpired
	}
	if challenge.AttemptCount >= challenge.MaxAttempts {
		if _, err := tx.Exec(ctx, `UPDATE otp_challenges SET consumed_at = $2 WHERE id = $1`, challenge.ID, now); err != nil {
			return VerifiedCustomerSession{}, fmt.Errorf("lock OTP challenge: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return VerifiedCustomerSession{}, fmt.Errorf("commit locked OTP challenge: %w", err)
		}
		return VerifiedCustomerSession{}, ErrOTPChallengeAttemptsExceeded
	}

	attempts := challenge.AttemptCount + 1
	if !hmac.Equal(challenge.OTPHash, otpHash) {
		if attempts >= challenge.MaxAttempts {
			_, err = tx.Exec(ctx, `
				UPDATE otp_challenges
				SET attempt_count = $2, consumed_at = $3
				WHERE id = $1`, challenge.ID, attempts, now)
		} else {
			_, err = tx.Exec(ctx, `
				UPDATE otp_challenges
				SET attempt_count = $2
				WHERE id = $1`, challenge.ID, attempts)
		}
		if err != nil {
			return VerifiedCustomerSession{}, fmt.Errorf("record invalid OTP attempt: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return VerifiedCustomerSession{}, fmt.Errorf("commit invalid OTP attempt: %w", err)
		}
		if attempts >= challenge.MaxAttempts {
			return VerifiedCustomerSession{}, ErrOTPChallengeAttemptsExceeded
		}
		return VerifiedCustomerSession{}, ErrOTPCodeInvalid
	}

	var customer Customer
	err = tx.QueryRow(ctx, `
		INSERT INTO customers (phone_number, phone_verified, is_active, last_login_at)
		VALUES ($1, TRUE, TRUE, $2)
		ON CONFLICT (phone_number) DO UPDATE SET
			phone_verified = TRUE,
			updated_at = $2,
			last_login_at = $2
		RETURNING id, phone_number, phone_verified, name, is_active, last_login_at`,
		phone, now,
	).Scan(
		&customer.ID,
		&customer.PhoneNumber,
		&customer.PhoneVerified,
		&customer.Name,
		&customer.IsActive,
		&customer.LastLoginAt,
	)
	if err != nil {
		return VerifiedCustomerSession{}, fmt.Errorf("upsert customer after OTP verification: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO customer_sessions (customer_id, session_token_hash, expires_at)
		VALUES ($1, $2, $3)`, customer.ID, sessionHash, sessionExpires); err != nil {
		return VerifiedCustomerSession{}, fmt.Errorf("create customer session: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE otp_challenges
		SET attempt_count = $2, verified_at = $3, consumed_at = $3
		WHERE id = $1 AND consumed_at IS NULL`, challenge.ID, attempts, now); err != nil {
		return VerifiedCustomerSession{}, fmt.Errorf("consume verified OTP challenge: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return VerifiedCustomerSession{}, fmt.Errorf("commit verified OTP transaction: %w", err)
	}
	return VerifiedCustomerSession{Customer: customer, SessionExpires: sessionExpires}, nil
}
