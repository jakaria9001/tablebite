package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

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
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := FromContext(r.Context())
			if !ok || user.Role != required && user.Role != "super_admin" {
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
