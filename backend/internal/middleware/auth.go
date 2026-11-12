package middleware

import (
	"context"
	"crypto/sha256"
	"net/http"
	"strings"
	"time"

	"github.com/jakaria9001/tablebite/internal/repository"
	"github.com/jakaria9001/tablebite/internal/session"
)

type contextKey string

const (
	userContextKey contextKey = "admin_user"
)

type AdminUser = session.AdminUser

func WithAdminUser(store session.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				if cookie, err := r.Cookie("tablebite_admin_session"); err == nil {
					token = cookie.Value
				}
			}
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			user, ok, err := store.Lookup(r.Context(), token)
			if err != nil || !ok || !user.Active || user.Expires.Before(time.Now()) {
				_ = store.Clear(r.Context(), token)
				http.SetCookie(w, &http.Cookie{Name: "tablebite_admin_session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteNoneMode})
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) string {
	const prefix = "Bearer "
	value := r.Header.Get("Authorization")
	if strings.HasPrefix(value, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(value, prefix))
	}
	return ""
}

func RequireRole(required string) func(http.Handler) http.Handler {
	return RequireRoles(required)
}

func RequireRoles(required ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := FromContext(r.Context())
			allowed := user.Role == "super_admin"
			for _, role := range required {
				if user.Role == role {
					allowed = true
					break
				}
			}
			if !ok || !allowed {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func FromContext(ctx context.Context) (AdminUser, bool) {
	user, ok := ctx.Value(userContextKey).(AdminUser)
	return user, ok
}

type WaiterUser = repository.Waiter

type WaiterSessionStore interface {
	Lookup(ctx context.Context, tokenHash []byte) (repository.Waiter, bool, error)
	Revoke(ctx context.Context, tokenHash []byte) error
}

const waiterSessionCookie = "tablebite_waiter_session"

func WithWaiterUser(store WaiterSessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(waiterSessionCookie)
			if err != nil || cookie.Value == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			hash := sha256.Sum256([]byte(cookie.Value))
			waiter, valid, err := store.Lookup(r.Context(), hash[:])
			if err != nil {
				http.Error(w, "authentication service unavailable", http.StatusServiceUnavailable)
				return
			}
			if !valid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), waiterContextKey, waiter)))
		})
	}
}

type waiterContextKeyType string

const waiterContextKey waiterContextKeyType = "waiter_user"

func WaiterFromContext(ctx context.Context) (repository.Waiter, bool) {
	waiter, ok := ctx.Value(waiterContextKey).(repository.Waiter)
	return waiter, ok
}

func WithStaffUser(waiterStore WaiterSessionStore, adminStore session.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token := adminTokenFromRequest(r); token != "" {
				admin, valid, err := adminStore.Lookup(r.Context(), token)
				if err != nil {
					http.Error(w, "authentication service unavailable", http.StatusServiceUnavailable)
					return
				}
				if valid && admin.Active && !admin.Expires.Before(time.Now()) &&
					(admin.Role == "admin" || admin.Role == "manager" || admin.Role == "super_admin") {
					next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey, admin)))
					return
				}
			}
			cookie, err := r.Cookie(waiterSessionCookie)
			if err != nil || cookie.Value == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			hash := sha256.Sum256([]byte(cookie.Value))
			waiter, valid, err := waiterStore.Lookup(r.Context(), hash[:])
			if err != nil {
				http.Error(w, "authentication service unavailable", http.StatusServiceUnavailable)
				return
			}
			if !valid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), waiterContextKey, waiter)))
		})
	}
}
