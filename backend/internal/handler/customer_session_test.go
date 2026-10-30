package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appmw "github.com/jakaria9001/tablebite/internal/middleware"
	"github.com/jakaria9001/tablebite/internal/repository"
)

type fakeCustomerSessionStore struct {
	customer repository.Customer
	valid    bool
	revoked  bool
}

func (s *fakeCustomerSessionStore) Lookup(context.Context, []byte) (repository.Customer, bool, error) {
	return s.customer, s.valid, nil
}

func (s *fakeCustomerSessionStore) Revoke(context.Context, []byte) error {
	s.revoked = true
	return nil
}

func TestMeReturnsSafeCustomerProfile(t *testing.T) {
	store := &fakeCustomerSessionStore{
		customer: repository.Customer{
			ID:            42,
			PhoneNumber:   "+919876543210",
			PhoneVerified: true,
			Name:          stringPointer("private name"),
			IsActive:      true,
		},
		valid: true,
	}
	authHandler := NewAuthHandler(nil, nil)
	authHandler.SetCustomerSessionStore(store)
	next := appmw.WithCustomerUser(store)(http.HandlerFunc(authHandler.Me))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "tablebite_customer_session", Value: "session-token"})
	res := httptest.NewRecorder()

	next.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	var body struct {
		Customer map[string]any `json:"customer"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Customer["id"] != float64(42) || body.Customer["phoneNumber"] != "+919876543210" || body.Customer["phoneVerified"] != true {
		t.Fatalf("customer response = %#v", body.Customer)
	}
	if _, exposed := body.Customer["name"]; exposed {
		t.Fatal("private customer name was exposed")
	}
}

func TestLogoutRevokesCustomerSessionAndExpiresCookie(t *testing.T) {
	store := &fakeCustomerSessionStore{valid: true}
	authHandler := NewAuthHandler(nil, nil)
	authHandler.SetCustomerSessionStore(store)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "tablebite_customer_session", Value: "session-token"})
	res := httptest.NewRecorder()

	authHandler.Logout(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNoContent)
	}
	if !store.revoked {
		t.Fatal("customer session was not revoked")
	}
	var found bool
	for _, cookie := range res.Result().Cookies() {
		if cookie.Name == "tablebite_customer_session" {
			found = true
			if cookie.MaxAge != -1 || cookie.HttpOnly == false {
				t.Fatalf("expired customer cookie = %#v", cookie)
			}
		}
	}
	if !found {
		t.Fatal("missing expired customer cookie")
	}
}

func TestLogoutWithoutCustomerSessionIsSafe(t *testing.T) {
	authHandler := NewAuthHandler(nil, nil)
	authHandler.SetCustomerSessionStore(&fakeCustomerSessionStore{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	res := httptest.NewRecorder()

	authHandler.Logout(res, req)

	if res.Code != http.StatusNoContent || strings.Contains(res.Body.String(), "panic") {
		t.Fatalf("logout response = %d %q", res.Code, res.Body.String())
	}
}

func stringPointer(value string) *string {
	return &value
}
