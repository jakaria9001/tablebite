package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWithAdminUserRejectsMissingSession(t *testing.T) {
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("next should not be called") })
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)

	WithAdminUser(next).ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
}

func TestRequireRoleAllowsAdminAndRejectsOtherRoles(t *testing.T) {
	cases := []struct {
		name string
		role string
		want int
	}{
		{name: "admin", role: "admin", want: http.StatusNoContent},
		{name: "super admin", role: "super_admin", want: http.StatusNoContent},
		{name: "manager", role: "manager", want: http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			req = req.WithContext(contextWithUser(req.Context(), AdminUser{Role: tc.role, Active: true, Expires: time.Now().Add(time.Hour)}))
			res := httptest.NewRecorder()

			RequireRole("admin")(next).ServeHTTP(res, req)

			if res.Code != tc.want {
				t.Fatalf("status = %d, want %d", res.Code, tc.want)
			}
		})
	}
}

func contextWithUser(ctx context.Context, user AdminUser) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}
