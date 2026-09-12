package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jakaria9001/tablebite/internal/auth"
	appmw "github.com/jakaria9001/tablebite/internal/middleware"
	"github.com/jakaria9001/tablebite/internal/model"
	"github.com/jakaria9001/tablebite/internal/repository"
	"github.com/jakaria9001/tablebite/internal/session"
)

type AuthHandler struct {
	admins   *repository.AdminRepository
	sessions session.Store
}

func NewAuthHandler(admins *repository.AdminRepository, sessions session.Store) *AuthHandler {
	return &AuthHandler{admins: admins, sessions: sessions}
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
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("tablebite_admin_session"); err == nil {
		_ = h.sessions.Clear(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "tablebite_admin_session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteNoneMode})
	http.SetCookie(w, &http.Cookie{Name: "tablebite_csrf", Value: "", Path: "/", MaxAge: -1, HttpOnly: false, Secure: true, SameSite: http.SameSiteNoneMode})
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) CSRFToken(w http.ResponseWriter, r *http.Request) {
	token, err := appmw.NewCSRFToken(os.Getenv("CSRF_SECRET"))
	if err != nil {
		http.Error(w, "could not create csrf token", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "tablebite_csrf", Value: token, Path: "/", HttpOnly: false, Secure: true, SameSite: http.SameSiteNoneMode, MaxAge: 3600})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"token": token})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := appmw.FromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"email": user.Email, "role": user.Role})
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
