package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/jakaria9001/tablebite/internal/auth"
	"github.com/jakaria9001/tablebite/internal/repository"
	"github.com/jakaria9001/tablebite/internal/session"
)

const customerSessionCookie = "tablebite_customer_session"

type customerContextKey string

const customerUserContextKey customerContextKey = "customer_user"

type CustomerUser = repository.Customer

type CustomerSessionStore interface {
	Lookup(ctx context.Context, tokenHash []byte) (repository.Customer, bool, error)
	Revoke(ctx context.Context, tokenHash []byte) error
}

func CustomerSessionToken(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(customerSessionCookie)
	if err != nil || cookie.Value == "" {
		return "", false
	}
	return cookie.Value, true
}

func CustomerSessionTokenHash(r *http.Request) ([]byte, bool) {
	token, ok := CustomerSessionToken(r)
	if !ok {
		return nil, false
	}
	return auth.HashSessionToken(token), true
}

func WithOptionalCustomer(store CustomerSessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hash, ok := CustomerSessionTokenHash(r)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			customer, valid, err := store.Lookup(r.Context(), hash)
			if err != nil {
				http.Error(w, "authentication service unavailable", http.StatusServiceUnavailable)
				return
			}
			if valid {
				r = r.WithContext(context.WithValue(r.Context(), customerUserContextKey, customer))
			}
			next.ServeHTTP(w, r)
		})
	}
}

func WithCustomerUser(store CustomerSessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hash, ok := CustomerSessionTokenHash(r)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			customer, valid, err := store.Lookup(r.Context(), hash)
			if err != nil {
				http.Error(w, "authentication service unavailable", http.StatusServiceUnavailable)
				return
			}
			if !valid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), customerUserContextKey, customer)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func WithCustomerOrAdmin(customerStore CustomerSessionStore, adminStore session.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := adminTokenFromRequest(r)
			if token != "" {
				admin, valid, err := adminStore.Lookup(r.Context(), token)
				if err != nil {
					http.Error(w, "authentication service unavailable", http.StatusServiceUnavailable)
					return
				}
				if valid && admin.Active && !admin.Expires.Before(time.Now()) {
					if admin.Role != "admin" && admin.Role != "super_admin" {
						http.Error(w, "forbidden", http.StatusForbidden)
						return
					}
					ctx := context.WithValue(r.Context(), userContextKey, admin)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			if hash, ok := CustomerSessionTokenHash(r); ok {
				customer, valid, err := customerStore.Lookup(r.Context(), hash)
				if err != nil {
					http.Error(w, "authentication service unavailable", http.StatusServiceUnavailable)
					return
				}
				if valid {
					ctx := context.WithValue(r.Context(), customerUserContextKey, customer)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		})
	}
}

func CustomerFromContext(ctx context.Context) (CustomerUser, bool) {
	customer, ok := ctx.Value(customerUserContextKey).(CustomerUser)
	return customer, ok
}

func adminTokenFromRequest(r *http.Request) string {
	if token := bearerToken(r); token != "" {
		return token
	}
	cookie, err := r.Cookie("tablebite_admin_session")
	if err != nil {
		return ""
	}
	return cookie.Value
}
