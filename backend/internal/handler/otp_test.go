package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jakaria9001/tablebite/internal/repository"
	"github.com/jakaria9001/tablebite/internal/service"
)

type fakeOTPRequester struct {
	err   error
	phone string
}

func (f *fakeOTPRequester) RequestOTP(_ context.Context, phone string) (service.OTPRequestResult, error) {
	f.phone = phone
	return service.OTPRequestResult{}, f.err
}

type fakeOTPRequestLimiter struct {
	err   error
	phone string
	ip    string
	now   time.Time
}

type fakeOTPVerifier struct {
	result service.OTPVerificationResult
	err    error
}

func (f *fakeOTPVerifier) VerifyOTPAndCreateSession(_ context.Context, _, _ string) (service.OTPVerificationResult, error) {
	return f.result, f.err
}

func (f *fakeOTPVerifier) RequestOTP(context.Context, string) (service.OTPRequestResult, error) {
	return service.OTPRequestResult{}, f.err
}

func (f *fakeOTPRequestLimiter) Allow(_ context.Context, phone, ip string, now time.Time) (repository.OTPRequestRateLimitDecision, error) {
	f.phone = phone
	f.ip = ip
	f.now = now
	return repository.OTPRequestRateLimitDecision{Allowed: f.err == nil}, f.err
}

func TestRequestOTPReturnsGenericSuccess(t *testing.T) {
	requester := &fakeOTPRequester{}
	limiter := &fakeOTPRequestLimiter{}
	handler := NewAuthHandlerWithOTP(nil, nil, requester, limiter)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/otp/request", strings.NewReader(`{"phone":"987-654-3210"}`))
	req.RemoteAddr = "203.0.113.10:43122"
	res := httptest.NewRecorder()

	handler.RequestOTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	var body map[string]string
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["message"] != "Verification code sent." {
		t.Fatalf("message = %q", body["message"])
	}
	if requester.phone != "+919876543210" || limiter.phone != "+919876543210" || limiter.ip != "203.0.113.10" {
		t.Fatalf("normalized request = phone %q, limiter phone %q, IP %q", requester.phone, limiter.phone, limiter.ip)
	}
}

func TestRequestOTPDoesNotRevealCustomerState(t *testing.T) {
	for name, requesterError := range map[string]error{
		"new customer":      nil,
		"existing customer": nil,
	} {
		t.Run(name, func(t *testing.T) {
			handler := NewAuthHandlerWithOTP(nil, nil, &fakeOTPRequester{err: requesterError}, &fakeOTPRequestLimiter{})
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"phone":"9876543210"}`))
			res := httptest.NewRecorder()

			handler.RequestOTP(res, req)

			if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), "Verification code sent.") {
				t.Fatalf("response = %d %q", res.Code, res.Body.String())
			}
		})
	}
}

func TestRequestOTPRejectsInvalidPhone(t *testing.T) {
	requester := &fakeOTPRequester{}
	limiter := &fakeOTPRequestLimiter{}
	handler := NewAuthHandlerWithOTP(nil, nil, requester, limiter)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"phone":"12345"}`))
	res := httptest.NewRecorder()

	handler.RequestOTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
	if requester.phone != "" || limiter.phone != "" {
		t.Fatal("invalid phone reached OTP service or limiter")
	}
}

func TestRequestOTPHandlesCooldownAndRateLimit(t *testing.T) {
	tests := []struct {
		name       string
		limiterErr error
		serviceErr error
		wantStatus int
	}{
		{name: "rate limited", limiterErr: repository.ErrOTPRequestRateLimited, wantStatus: http.StatusTooManyRequests},
		{name: "cooldown", serviceErr: service.ErrOTPResendCooldown, wantStatus: http.StatusTooManyRequests},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewAuthHandlerWithOTP(nil, nil, &fakeOTPRequester{err: tc.serviceErr}, &fakeOTPRequestLimiter{err: tc.limiterErr})
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"phone":"9876543210"}`))
			res := httptest.NewRecorder()

			handler.RequestOTP(res, req)

			if res.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", res.Code, tc.wantStatus)
			}
		})
	}
}

func TestRequestOTPHandlesProviderFailureGenerically(t *testing.T) {
	handler := NewAuthHandlerWithOTP(nil, nil, &fakeOTPRequester{err: errors.New("provider failure")}, &fakeOTPRequestLimiter{})
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"phone":"9876543210"}`))
	res := httptest.NewRecorder()

	handler.RequestOTP(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusServiceUnavailable)
	}
	if strings.Contains(res.Body.String(), "provider failure") {
		t.Fatalf("provider error leaked: %q", res.Body.String())
	}
}

func TestVerifyOTPSetsHTTPOnlyCustomerCookieAndReturnsCustomer(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	expires := time.Now().Add(time.Hour)
	authHandler := NewAuthHandlerWithOTP(nil, nil, &fakeOTPVerifier{
		result: service.OTPVerificationResult{
			Customer: repository.Customer{
				ID:            42,
				PhoneNumber:   "+919876543210",
				PhoneVerified: true,
			},
			SessionToken:   "raw-token-only-for-cookie",
			SessionExpires: expires,
		},
	}, nil)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"phone":"9876543210","otp":"123456"}`))
	res := httptest.NewRecorder()

	authHandler.VerifyOTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	cookie := res.Result().Cookies()[0]
	if cookie.Name != "tablebite_customer_session" || cookie.Value != "raw-token-only-for-cookie" {
		t.Fatalf("session cookie = %#v", cookie)
	}
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteNoneMode || cookie.Path != "/" {
		t.Fatalf("cookie security attributes = %#v", cookie)
	}
	if strings.Contains(res.Body.String(), "raw-token-only-for-cookie") {
		t.Fatal("raw session token leaked in response body")
	}
	if !strings.Contains(res.Body.String(), `"id":42`) {
		t.Fatalf("customer response = %q", res.Body.String())
	}
}

func TestVerifyOTPReturnsGenericErrorForInvalidCode(t *testing.T) {
	authHandler := NewAuthHandlerWithOTP(nil, nil, &fakeOTPVerifier{err: service.ErrInvalidOTP}, nil)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"phone":"9876543210","otp":"000000"}`))
	res := httptest.NewRecorder()

	authHandler.VerifyOTP(res, req)

	if res.Code != http.StatusBadRequest || !strings.Contains(res.Body.String(), "invalid or expired") {
		t.Fatalf("response = %d %q", res.Code, res.Body.String())
	}
}
