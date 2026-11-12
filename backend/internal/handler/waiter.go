package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jakaria9001/tablebite/internal/auth"
	appmw "github.com/jakaria9001/tablebite/internal/middleware"
	"github.com/jakaria9001/tablebite/internal/repository"
	"github.com/jakaria9001/tablebite/internal/service"
)

type WaiterStore interface {
	Provision(ctx context.Context, phone, name string) (repository.Waiter, error)
	FindByPhone(ctx context.Context, phone string) (repository.Waiter, error)
	CreateSession(ctx context.Context, waiterID int64, tokenHash []byte, expiresAt time.Time) error
	Lookup(ctx context.Context, tokenHash []byte) (repository.Waiter, bool, error)
	Revoke(ctx context.Context, tokenHash []byte) error
	ListOrders(ctx context.Context) ([]repository.Order, error)
	GetOrder(ctx context.Context, orderID int64) (repository.Order, error)
	TransitionOrder(ctx context.Context, orderID int64, newStatus string, waiterID, adminID *int64, note string) (repository.Order, error)
}

type WaiterHandler struct {
	store        WaiterStore
	otpRequester OTPRequester
	otpVerifier  OTPVerifier
}

func NewWaiterHandler(store WaiterStore, requester OTPRequester, verifier OTPVerifier) *WaiterHandler {
	return &WaiterHandler{store: store, otpRequester: requester, otpVerifier: verifier}
}

func (h *WaiterHandler) Provision(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Phone string `json:"phone"`
		Name  string `json:"name"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil || strings.TrimSpace(payload.Name) == "" {
		respondError(w, r, http.StatusBadRequest, "invalid_waiter", "A phone number and name are required.")
		return
	}
	phone, err := service.NormalizePhone(payload.Phone)
	if err != nil {
		respondError(w, r, http.StatusBadRequest, "invalid_phone", "Enter a valid Indian mobile number.")
		return
	}
	waiter, err := h.store.Provision(r.Context(), phone, strings.TrimSpace(payload.Name))
	if errors.Is(err, repository.ErrCustomerNotVerified) {
		respondError(w, r, http.StatusConflict, "customer_not_verified", "The customer must verify their phone before becoming a waiter.")
		return
	}
	if err != nil {
		respondError(w, r, http.StatusServiceUnavailable, "waiter_unavailable", "We could not add this waiter.")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"waiter": waiter})
}

func (h *WaiterHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	if h.otpRequester == nil {
		respondError(w, r, http.StatusServiceUnavailable, "otp_unavailable", "Verification is temporarily unavailable.")
		return
	}
	var payload struct {
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, r, http.StatusBadRequest, "invalid_phone", "Enter a valid phone number.")
		return
	}
	phone, err := service.NormalizePhone(payload.Phone)
	if err != nil {
		respondError(w, r, http.StatusBadRequest, "invalid_phone", "Enter a valid Indian mobile number.")
		return
	}
	if _, err := h.store.FindByPhone(r.Context(), phone); errors.Is(err, repository.ErrWaiterNotFound) {
		respondError(w, r, http.StatusUnauthorized, "waiter_not_found", "Waiter access is not enabled for this number.")
		return
	} else if err != nil {
		respondError(w, r, http.StatusServiceUnavailable, "waiter_unavailable", "We could not check waiter access.")
		return
	}
	if _, err := h.otpRequester.RequestOTP(r.Context(), phone); err != nil {
		if errors.Is(err, service.ErrOTPResendCooldown) {
			respondError(w, r, http.StatusTooManyRequests, "otp_cooldown", "Please wait before requesting another code.")
			return
		}
		respondError(w, r, http.StatusServiceUnavailable, "otp_unavailable", "We could not send a verification code.")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Verification code sent."})
}

func (h *WaiterHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	if h.otpVerifier == nil {
		respondError(w, r, http.StatusServiceUnavailable, "otp_unavailable", "Verification is temporarily unavailable.")
		return
	}
	var payload struct {
		Phone string `json:"phone"`
		OTP   string `json:"otp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, r, http.StatusBadRequest, "invalid_otp", "Enter a valid verification code.")
		return
	}
	phone, err := service.NormalizePhone(payload.Phone)
	if err != nil {
		respondError(w, r, http.StatusBadRequest, "invalid_otp", "The verification code is invalid or expired.")
		return
	}
	result, err := h.otpVerifier.VerifyOTPAndCreateSession(r.Context(), phone, payload.OTP)
	if err != nil {
		respondError(w, r, http.StatusBadRequest, "invalid_otp", "The verification code is invalid or expired.")
		return
	}
	waiter, err := h.store.FindByPhone(r.Context(), phone)
	if err != nil {
		respondError(w, r, http.StatusUnauthorized, "waiter_not_found", "Waiter access is not enabled for this number.")
		return
	}
	token, err := auth.GenerateSessionToken()
	if err != nil {
		respondError(w, r, http.StatusServiceUnavailable, "session_unavailable", "We could not create a waiter session.")
		return
	}
	if err := h.store.CreateSession(r.Context(), waiter.ID, auth.HashSessionToken(token), result.SessionExpires); err != nil {
		respondError(w, r, http.StatusServiceUnavailable, "session_unavailable", "We could not create a waiter session.")
		return
	}
	secure := customerCookieSecure()
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, &http.Cookie{
		Name: "tablebite_waiter_session", Value: token, Path: "/", HttpOnly: true,
		Secure: secure, SameSite: sameSite, Expires: result.SessionExpires,
		MaxAge: maxAgeUntil(result.SessionExpires),
	})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"waiter": waiter})
}

func (h *WaiterHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("tablebite_waiter_session"); err == nil {
		_ = h.store.Revoke(r.Context(), auth.HashSessionToken(cookie.Value))
	}
	http.SetCookie(w, &http.Cookie{Name: "tablebite_waiter_session", Value: "", Path: "/", MaxAge: -1, Expires: time.Unix(1, 0), HttpOnly: true, Secure: customerCookieSecure(), SameSite: http.SameSiteLaxMode})
	w.WriteHeader(http.StatusNoContent)
}

func (h *WaiterHandler) Orders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.store.ListOrders(r.Context())
	if err != nil {
		respondError(w, r, http.StatusServiceUnavailable, "orders_unavailable", "We could not load incoming orders.")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"orders": orders})
}

func (h *WaiterHandler) Order(w http.ResponseWriter, r *http.Request) {
	orderID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orderID <= 0 {
		respondError(w, r, http.StatusNotFound, "order_not_found", "Order not found.")
		return
	}
	order, err := h.store.GetOrder(r.Context(), orderID)
	if errors.Is(err, repository.ErrOrderNotFound) {
		respondError(w, r, http.StatusNotFound, "order_not_found", "Order not found.")
		return
	}
	if err != nil {
		respondError(w, r, http.StatusServiceUnavailable, "orders_unavailable", "We could not load this order.")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"order": order})
}

func (h *WaiterHandler) Transition(w http.ResponseWriter, r *http.Request) {
	orderID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orderID <= 0 {
		respondError(w, r, http.StatusNotFound, "order_not_found", "Order not found.")
		return
	}
	var payload struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, r, http.StatusBadRequest, "invalid_status", "Enter a valid order status.")
		return
	}
	var waiterID, adminID *int64
	if waiter, ok := appmw.WaiterFromContext(r.Context()); ok {
		waiterID = &waiter.ID
	}
	if admin, ok := appmw.FromContext(r.Context()); ok {
		adminID = &admin.ID
	}
	order, err := h.store.TransitionOrder(r.Context(), orderID, strings.ToLower(strings.TrimSpace(payload.Status)), waiterID, adminID, strings.TrimSpace(payload.Note))
	switch {
	case errors.Is(err, repository.ErrOrderNotFound):
		respondError(w, r, http.StatusNotFound, "order_not_found", "Order not found.")
	case errors.Is(err, repository.ErrOrderStatusConflict):
		respondError(w, r, http.StatusConflict, "status_conflict", "This order has already changed.")
	case errors.Is(err, repository.ErrOrderStatusInvalid):
		respondError(w, r, http.StatusBadRequest, "invalid_status", "That status transition is not allowed.")
	case err != nil:
		respondError(w, r, http.StatusServiceUnavailable, "orders_unavailable", "We could not update this order.")
	default:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"order": order})
	}
}
