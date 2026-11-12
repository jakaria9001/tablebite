package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jakaria9001/tablebite/internal/repository"
)

type waiterAuthStore struct {
	waiter repository.Waiter
	valid  bool
}

func (s waiterAuthStore) Lookup(context.Context, []byte) (repository.Waiter, bool, error) {
	return s.waiter, s.valid, nil
}

func (waiterAuthStore) Revoke(context.Context, []byte) error {
	return nil
}

func TestWithWaiterUserAcceptsWaiterCookie(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		waiter, ok := WaiterFromContext(r.Context())
		if !ok || waiter.ID != 7 {
			t.Fatal("waiter missing from context")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: waiterSessionCookie, Value: "waiter-token"})
	res := httptest.NewRecorder()

	WithWaiterUser(waiterAuthStore{waiter: repository.Waiter{ID: 7, Active: true}, valid: true})(next).ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNoContent)
	}
}

func TestWithWaiterUserRejectsMissingCookie(t *testing.T) {
	res := httptest.NewRecorder()
	WithWaiterUser(waiterAuthStore{})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next should not be called")
	})).ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
}
