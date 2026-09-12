package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type contextKey string

const (
	userContextKey contextKey = "admin_user"
)

type AdminUser struct {
	ID      int64
	Email   string
	Role    string
	Active  bool
	Expires time.Time
}

var (
	sessionMu sync.RWMutex
	sessions  = map[string]AdminUser{}
)

func WithAdminUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("tablebite_admin_session")
		if err != nil || cookie.Value == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		user, ok := lookupSession(cookie.Value)
		if !ok || !user.Active || user.Expires.Before(time.Now()) {
			ClearSession(cookie.Value)
			http.SetCookie(w, &http.Cookie{Name: "tablebite_admin_session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteNoneMode})
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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

func RegisterSession(token string, user AdminUser) {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	sessions[token] = user
}

func ClearSession(token string) {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	delete(sessions, token)
}

func lookupSession(token string) (AdminUser, bool) {
	if token == "" {
		return AdminUser{}, false
	}
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	user, ok := sessions[token]
	if !ok {
		return AdminUser{}, false
	}
	return user, true
}
