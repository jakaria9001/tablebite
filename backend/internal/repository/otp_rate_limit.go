package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrOTPRequestRateLimited = errors.New("OTP request rate limit exceeded")

type OTPRequestRateLimitConfig struct {
	Window     time.Duration
	PhoneLimit int
	IPLimit    int
}

func LoadOTPRequestRateLimitConfig() (OTPRequestRateLimitConfig, error) {
	window, err := positiveDurationFromEnv("OTP_RATE_LIMIT_WINDOW_SECONDS", time.Hour)
	if err != nil {
		return OTPRequestRateLimitConfig{}, err
	}
	phoneLimit, err := positiveIntFromEnv("OTP_PHONE_RATE_LIMIT", 5)
	if err != nil {
		return OTPRequestRateLimitConfig{}, err
	}
	ipLimit, err := positiveIntFromEnv("OTP_IP_RATE_LIMIT", 20)
	if err != nil {
		return OTPRequestRateLimitConfig{}, err
	}
	return OTPRequestRateLimitConfig{Window: window, PhoneLimit: phoneLimit, IPLimit: ipLimit}, nil
}

type OTPRequestRateLimitDecision struct {
	Allowed    bool
	RetryAfter time.Duration
}

type OTPRequestRateLimiter struct {
	db     *pgxpool.Pool
	config OTPRequestRateLimitConfig
}

func NewOTPRequestRateLimiter(db *pgxpool.Pool, config OTPRequestRateLimitConfig) *OTPRequestRateLimiter {
	return &OTPRequestRateLimiter{db: db, config: config}
}

type rateLimitRule struct {
	key   string
	limit int
}

func (l *OTPRequestRateLimiter) Allow(ctx context.Context, phone, ip string, now time.Time) (OTPRequestRateLimitDecision, error) {
	if l.config.Window <= 0 || l.config.PhoneLimit <= 0 || l.config.IPLimit <= 0 {
		return OTPRequestRateLimitDecision{}, errors.New("invalid OTP request rate limit configuration")
	}

	rules := []rateLimitRule{
		{key: "ip:" + ip, limit: l.config.IPLimit},
		{key: "phone:" + phone, limit: l.config.PhoneLimit},
	}
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].key < rules[j].key
	})

	tx, err := l.db.Begin(ctx)
	if err != nil {
		return OTPRequestRateLimitDecision{}, fmt.Errorf("begin OTP rate limit transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	cutoff := now.Add(-l.config.Window)
	for _, rule := range rules {
		if _, err := tx.Exec(ctx, `
			INSERT INTO otp_request_rate_limits(rate_key, window_started_at, request_count, last_requested_at)
			VALUES ($1, $2, 1, $2)
			ON CONFLICT (rate_key) DO NOTHING`, rule.key, now); err != nil {
			return OTPRequestRateLimitDecision{}, fmt.Errorf("initialize OTP rate limit: %w", err)
		}

		var windowStartedAt time.Time
		var requestCount int
		err := tx.QueryRow(ctx, `
			SELECT window_started_at, request_count
			FROM otp_request_rate_limits
			WHERE rate_key = $1
			FOR UPDATE`, rule.key,
		).Scan(&windowStartedAt, &requestCount)
		if err != nil {
			return OTPRequestRateLimitDecision{}, fmt.Errorf("load OTP rate limit: %w", err)
		}

		if windowStartedAt.After(cutoff) && requestCount >= rule.limit {
			retryAfter := windowStartedAt.Add(l.config.Window).Sub(now)
			if retryAfter < 0 {
				retryAfter = 0
			}
			return OTPRequestRateLimitDecision{RetryAfter: retryAfter}, ErrOTPRequestRateLimited
		}

		if windowStartedAt.After(cutoff) {
			requestCount++
		} else {
			windowStartedAt = now
			requestCount = 1
		}

		if _, err := tx.Exec(ctx, `
			UPDATE otp_request_rate_limits
			SET window_started_at = $2, request_count = $3, last_requested_at = $4
			WHERE rate_key = $1`, rule.key, windowStartedAt, requestCount, now); err != nil {
			return OTPRequestRateLimitDecision{}, fmt.Errorf("update OTP rate limit: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return OTPRequestRateLimitDecision{}, fmt.Errorf("commit OTP rate limit transaction: %w", err)
	}
	return OTPRequestRateLimitDecision{Allowed: true}, nil
}

func (l *OTPRequestRateLimiter) Cleanup(ctx context.Context, before time.Time) error {
	_, err := l.db.Exec(ctx, `DELETE FROM otp_request_rate_limits WHERE window_started_at < $1`, before)
	if err != nil {
		return fmt.Errorf("cleanup OTP rate limits: %w", err)
	}
	return nil
}

func positiveDurationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
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
