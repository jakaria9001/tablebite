package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jakaria9001/tablebite/internal/auth"
	appmw "github.com/jakaria9001/tablebite/internal/middleware"
	"github.com/jakaria9001/tablebite/internal/model"
	"github.com/jakaria9001/tablebite/internal/repository"
	"github.com/jakaria9001/tablebite/internal/service"
	"github.com/jakaria9001/tablebite/internal/session"
)

type AuthHandler struct {
	admins           *repository.AdminRepository
	sessions         session.Store
	otpService       OTPRequester
	otpLimiter       OTPRequestLimiter
	otpVerifier      OTPVerifier
	customerSessions appmw.CustomerSessionStore
}

func NewAuthHandler(admins *repository.AdminRepository, sessions session.Store) *AuthHandler {
	return &AuthHandler{admins: admins, sessions: sessions}
}

type OTPRequester interface {
	RequestOTP(ctx context.Context, phone string) (service.OTPRequestResult, error)
}

type OTPVerifier interface {
	VerifyOTPAndCreateSession(ctx context.Context, phone, code string) (service.OTPVerificationResult, error)
}

type OTPRequestLimiter interface {
	Allow(ctx context.Context, phone, ip string, now time.Time) (repository.OTPRequestRateLimitDecision, error)
}

func NewAuthHandlerWithOTP(admins *repository.AdminRepository, sessions session.Store, otpService OTPRequester, otpLimiter OTPRequestLimiter) *AuthHandler {
	return &AuthHandler{
		admins:     admins,
		sessions:   sessions,
		otpService: otpService,
		otpLimiter: otpLimiter,
		otpVerifier: func() OTPVerifier {
			verifier, _ := otpService.(OTPVerifier)
			return verifier
		}(),
	}
}

func (h *AuthHandler) SetCustomerSessionStore(store appmw.CustomerSessionStore) {
	h.customerSessions = store
}

func (h *AuthHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	if h.otpService == nil || h.otpLimiter == nil {
		http.Error(w, "verification service unavailable", http.StatusServiceUnavailable)
		return
	}

	var payload struct {
		Phone string `json:"phone"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		writeOTPError(w, http.StatusBadRequest, "Please enter a valid phone number.")
		return
	}

	phone, err := service.NormalizePhone(payload.Phone)
	if err != nil {
		writeOTPError(w, http.StatusBadRequest, "Please enter a valid Indian mobile number.")
		return
	}

	decision, err := h.otpLimiter.Allow(r.Context(), phone, requestIP(r), time.Now())
	if errors.Is(err, repository.ErrOTPRequestRateLimited) {
		if decision.RetryAfter > 0 {
			w.Header().Set("Retry-After", strconv.FormatInt(int64(decision.RetryAfter.Seconds()+0.999), 10))
		}
		writeOTPError(w, http.StatusTooManyRequests, "Please wait before requesting another code.")
		return
	}
	if err != nil {
		http.Error(w, "verification service unavailable", http.StatusServiceUnavailable)
		return
	}

	if _, err := h.otpService.RequestOTP(r.Context(), phone); err != nil {
		if errors.Is(err, service.ErrOTPResendCooldown) {
			writeOTPError(w, http.StatusTooManyRequests, "Please wait before requesting another code.")
			return
		}
		if errors.Is(err, service.ErrInvalidPhone) {
			writeOTPError(w, http.StatusBadRequest, "Please enter a valid Indian mobile number.")
			return
		}
		http.Error(w, "We could not send a verification code. Please try again.", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Verification code sent."})
}

func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	if h.otpVerifier == nil {
		http.Error(w, "verification service unavailable", http.StatusServiceUnavailable)
		return
	}

	var payload struct {
		Phone string `json:"phone"`
		OTP   string `json:"otp"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		writeOTPError(w, http.StatusBadRequest, "Please enter a valid verification code.")
		return
	}

	result, err := h.otpVerifier.VerifyOTPAndCreateSession(r.Context(), payload.Phone, payload.OTP)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPhone), errors.Is(err, service.ErrInvalidOTP),
			errors.Is(err, service.ErrNoOTPChallenge), errors.Is(err, service.ErrOTPExpired),
			errors.Is(err, service.ErrOTPConsumed):
			writeOTPError(w, http.StatusBadRequest, "The verification code is invalid or expired.")
		case errors.Is(err, service.ErrOTPAttemptsExceeded):
			writeOTPError(w, http.StatusTooManyRequests, "Too many verification attempts. Request a new code.")
		default:
			http.Error(w, "We could not verify the code. Please try again.", http.StatusServiceUnavailable)
		}
		return
	}

	secure := customerCookieSecure()
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "tablebite_customer_session",
		Value:    result.SessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Expires:  result.SessionExpires,
		MaxAge:   maxAgeUntil(result.SessionExpires),
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"customer": map[string]any{
			"id":             result.Customer.ID,
			"phone":          result.Customer.PhoneNumber,
			"phone_verified": result.Customer.PhoneVerified,
			"name":           result.Customer.Name,
		},
	})
}

func writeOTPError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func requestIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	if value := strings.TrimSpace(r.RemoteAddr); value != "" {
		return value
	}
	return "unknown"
}

func customerCookieSecure() bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(os.Getenv("COOKIE_SECURE")), "true")
}

func maxAgeUntil(expires time.Time) int {
	seconds := int(time.Until(expires).Seconds())
	if seconds < 1 {
		return 1
	}
	return seconds
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	email := auth.NormalizeEmail(payload.Email)
	password := strings.TrimSpace(payload.Password)
	if email == "" || password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	admin, err := h.authenticate(r.Context(), email, password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := auth.GenerateSessionToken()
	if err != nil {
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}

	sessionUser := session.AdminUser{ID: admin.ID, Email: admin.Email, Role: admin.Role, Active: true, Expires: time.Now().Add(auth.SessionExpiry)}
	if err := h.sessions.Register(r.Context(), token, sessionUser); err != nil {
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "tablebite_admin_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   int(auth.SessionExpiry.Seconds()),
		Expires:  time.Now().Add(auth.SessionExpiry),
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "token": token})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if hash, ok := appmw.CustomerSessionTokenHash(r); ok && h.customerSessions != nil {
		_ = h.customerSessions.Revoke(r.Context(), hash)
	}
	if token := bearerToken(r); token != "" {
		_ = h.sessions.Clear(r.Context(), token)
	}
	if cookie, err := r.Cookie("tablebite_admin_session"); err == nil {
		_ = h.sessions.Clear(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "tablebite_admin_session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteNoneMode})
	csrfCookie := &http.Cookie{Name: "tablebite_csrf", Value: "", Path: "/", MaxAge: -1, HttpOnly: false}
	if customerCookieSecure() {
		csrfCookie.Secure = true
		csrfCookie.SameSite = http.SameSiteNoneMode
	} else {
		csrfCookie.SameSite = http.SameSiteLaxMode
	}
	http.SetCookie(w, csrfCookie)
	http.SetCookie(w, customerSessionCookieExpired())
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) CSRFToken(w http.ResponseWriter, r *http.Request) {
	token, err := appmw.NewCSRFToken(os.Getenv("CSRF_SECRET"))
	if err != nil {
		http.Error(w, "could not create csrf token", http.StatusInternalServerError)
		return
	}
	sameSite := http.SameSiteLaxMode
	secure := customerCookieSecure()
	if secure {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, &http.Cookie{Name: "tablebite_csrf", Value: token, Path: "/", HttpOnly: false, Secure: secure, SameSite: sameSite, MaxAge: 3600})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"token": token})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if customer, ok := appmw.CustomerFromContext(r.Context()); ok {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"customer": map[string]any{
				"id":            customer.ID,
				"phoneNumber":   customer.PhoneNumber,
				"phoneVerified": customer.PhoneVerified,
			},
		})
		return
	}
	user, ok := appmw.FromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"email": user.Email, "role": user.Role})
}

func customerSessionCookieExpired() *http.Cookie {
	secure := customerCookieSecure()
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}
	return &http.Cookie{
		Name:     "tablebite_customer_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
	}
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user, ok := appmw.FromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if payload.NewPassword == "" {
		http.Error(w, "new password is required", http.StatusBadRequest)
		return
	}
	if err := auth.ValidatePassword(payload.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	admin, err := h.authenticate(r.Context(), user.Email, payload.CurrentPassword)
	if err != nil {
		http.Error(w, "current password is invalid", http.StatusUnauthorized)
		return
	}

	hash, err := auth.HashPassword(payload.NewPassword)
	if err != nil {
		http.Error(w, "could not change password", http.StatusInternalServerError)
		return
	}
	if err := h.admins.UpdatePasswordHash(r.Context(), admin.ID, hash); err != nil {
		http.Error(w, "could not change password", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) authenticate(ctx context.Context, email, password string) (model.Admin, error) {
	admin, err := h.admins.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrAdminNotFound) {
			return model.Admin{}, auth.ErrInvalidCredentials
		}
		return model.Admin{}, err
	}
	if !admin.IsActive {
		return model.Admin{}, auth.ErrInvalidCredentials
	}
	if err := auth.ComparePassword(admin.PasswordHash, password); err != nil {
		return model.Admin{}, auth.ErrInvalidCredentials
	}
	return admin, nil
}

func bearerToken(r *http.Request) string {
	value := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if strings.HasPrefix(value, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(value, prefix))
	}
	return ""
}
