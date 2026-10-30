package notification

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
)

var ErrMockOTPDisabled = errors.New("mock OTP provider is disabled in production")

type OTPProvider interface {
	SendOTP(ctx context.Context, phone string, code string) error
}

type MockOTPProvider struct {
	mu        sync.Mutex
	LastCode  string
	LastPhone string
	Logger    *slog.Logger
}

func NewMockOTPProvider(logger *slog.Logger) *MockOTPProvider {
	if logger == nil {
		logger = slog.Default()
	}
	return &MockOTPProvider{Logger: logger}
}

func (p *MockOTPProvider) SendOTP(_ context.Context, phone string, code string) error {
	p.mu.Lock()
	p.LastPhone = phone
	p.LastCode = code
	p.mu.Unlock()
	p.Logger.Warn("DEVELOPMENT ONLY: mock OTP generated", "phone", phone, "otp", code)
	return nil
}

func NewOTPProvider(provider, appEnv string) (OTPProvider, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "mock":
		if strings.EqualFold(strings.TrimSpace(appEnv), "production") {
			return nil, ErrMockOTPDisabled
		}
		return NewMockOTPProvider(nil), nil
	default:
		return nil, errors.New("unsupported OTP provider")
	}
}
