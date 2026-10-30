package repository

import (
	"testing"
	"time"
)

func TestLoadOTPRequestRateLimitConfigUsesDefaults(t *testing.T) {
	for _, key := range []string{"OTP_RATE_LIMIT_WINDOW_SECONDS", "OTP_PHONE_RATE_LIMIT", "OTP_IP_RATE_LIMIT"} {
		t.Setenv(key, "")
	}

	config, err := LoadOTPRequestRateLimitConfig()
	if err != nil {
		t.Fatalf("load defaults: %v", err)
	}
	if config.Window != time.Hour || config.PhoneLimit != 5 || config.IPLimit != 20 {
		t.Fatalf("config = %#v", config)
	}
}

func TestLoadOTPRequestRateLimitConfigRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		key   string
		value string
	}{
		{key: "OTP_RATE_LIMIT_WINDOW_SECONDS", value: "0"},
		{key: "OTP_PHONE_RATE_LIMIT", value: "-1"},
		{key: "OTP_IP_RATE_LIMIT", value: "not-a-number"},
	}
	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			t.Setenv("OTP_RATE_LIMIT_WINDOW_SECONDS", "")
			t.Setenv("OTP_PHONE_RATE_LIMIT", "")
			t.Setenv("OTP_IP_RATE_LIMIT", "")
			t.Setenv(test.key, test.value)
			if _, err := LoadOTPRequestRateLimitConfig(); err == nil {
				t.Fatal("expected invalid configuration error")
			}
		})
	}
}
