package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jakaria9001/tablebite/internal/repository"
	"github.com/jakaria9001/tablebite/internal/session"
)

type fakeCustomerSessionStore struct {
	customer repository.Customer
	valid    bool
	err      error
	hash     []byte
	revoked  bool
}

func (s *fakeCustomerSessionStore) Lookup(_ context.Context, hash []byte) (repository.Customer, bool, error) {
	s.hash = append([]byte(nil), hash...)
	return s.customer, s.valid, s.err
}

func (s *fakeCustomerSessionStore) Revoke(_ context.Context, hash []byte) error {
	s.hash = append([]byte(nil), hash...)
	s.revoked = true
	return nil
}

type fakeAdminSessionStore struct {
	user  session.AdminUser
	valid bool
}

func (s *fakeAdminSessionStore) Register(context.Context, string, session.AdminUser) error {
	return nil
}

func (s *fakeAdminSessionStore) Lookup(context.Context, string) (session.AdminUser, bool, error) {
	return s.user, s.valid, nil
}

func (s *fakeAdminSessionStore) Clear(context.Context, string) error {
	return nil
}

func TestWithCustomerUserAcceptsValidSession(t *testing.T) {
	store := &fakeCustomerSessionStore{
		customer: repository.Customer{ID: 42, PhoneNumber: "+919876543210", IsActive: true},
		valid:    true,
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := CustomerFromContext(r.Context()); !ok {
			t.Fatal("customer missing from context")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: customerSessionCookie, Value: "customer-token"})
	res := httptest.NewRecorder()

	WithCustomerUser(store)(next).ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNoContent)
	}
	if len(store.hash) == 0 {
		t.Fatal("session token was not hashed")
	}
}

func TestWithCustomerUserRejectsInvalidExpiredOrRevokedSessions(t *testing.T) {
	for _, name := range []string{"invalid", "expired", "revoked"} {
		t.Run(name, func(t *testing.T) {
			store := &fakeCustomerSessionStore{}
			next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Fatal("next should not be called")
			})
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.AddCookie(&http.Cookie{Name: customerSessionCookie, Value: name + "-token"})
			res := httptest.NewRecorder()

			WithCustomerUser(store)(next).ServeHTTP(res, req)

			if res.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestWithCustomerUserRejectsMissingSession(t *testing.T) {
	store := &fakeCustomerSessionStore{}
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next should not be called")
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	WithCustomerUser(store)(next).ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
}

func TestWithOptionalCustomerDoesNotAuthenticateAdminCookie(t *testing.T) {
	store := &fakeCustomerSessionStore{}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := CustomerFromContext(r.Context()); ok {
			t.Fatal("admin cookie authenticated as customer")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "tablebite_admin_session", Value: "admin-token"})
	res := httptest.NewRecorder()

	WithOptionalCustomer(store)(next).ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNoContent)
	}
}

func TestWithCustomerOrAdminPrefersAdminAndPreservesRoleAuthorization(t *testing.T) {
	customerStore := &fakeCustomerSessionStore{
		customer: repository.Customer{ID: 42, PhoneNumber: "+919876543210", IsActive: true},
		valid:    true,
	}
	adminStore := &fakeAdminSessionStore{
		user:  session.AdminUser{ID: 7, Role: "admin", Active: true, Expires: time.Now().Add(time.Hour)},
		valid: true,
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := FromContext(r.Context()); !ok {
			t.Fatal("admin missing from context")
		}
		if _, ok := CustomerFromContext(r.Context()); ok {
			t.Fatal("admin request authenticated as customer")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	req.AddCookie(&http.Cookie{Name: customerSessionCookie, Value: "customer-token"})
	res := httptest.NewRecorder()

	WithCustomerOrAdmin(customerStore, adminStore)(next).ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNoContent)
	}
}

func TestWithCustomerOrAdminRejectsManagerRole(t *testing.T) {
	adminStore := &fakeAdminSessionStore{
		user:  session.AdminUser{Role: "manager", Active: true, Expires: time.Now().Add(time.Hour)},
		valid: true,
	}
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("manager should not be authorized")
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer manager-token")
	res := httptest.NewRecorder()

	WithCustomerOrAdmin(&fakeCustomerSessionStore{}, adminStore)(next).ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusForbidden)
	}
}
