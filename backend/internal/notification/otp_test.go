package notification

import (
	"context"
	"errors"
	"testing"
)

func TestNewOTPProviderDisablesMockInProduction(t *testing.T) {
	_, err := NewOTPProvider("mock", "production")
	if !errors.Is(err, ErrMockOTPDisabled) {
		t.Fatalf("production mock provider error = %v, want %v", err, ErrMockOTPDisabled)
	}
}

func TestMockOTPProviderCapturesDevelopmentCode(t *testing.T) {
	provider := NewMockOTPProvider(nil)
	if err := provider.SendOTP(context.Background(), "+919876543210", "123456"); err != nil {
		t.Fatalf("send mock OTP: %v", err)
	}
	if provider.LastPhone != "+919876543210" || provider.LastCode != "123456" {
		t.Fatalf("mock provider captured (%q, %q)", provider.LastPhone, provider.LastCode)
	}
}
