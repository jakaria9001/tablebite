package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chimw "github.com/go-chi/chi/v5/middleware"
	appmw "github.com/jakaria9001/tablebite/internal/middleware"
)

func TestHealthReturnsJSONAndRequestID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	res := httptest.NewRecorder()

	chimw.RequestID(http.HandlerFunc(Health)).ServeHTTP(res, req)

	if res.Code != http.StatusOK || res.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("health response = %d %q", res.Code, res.Header().Get("Content-Type"))
	}
	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if body["status"] != "ok" || body["request_id"] == "" {
		t.Fatalf("unexpected health response: %#v", body)
	}
}

func TestLoginRejectsInvalidJSONWithAPIErrorShape(t *testing.T) {
	handler := NewAuthHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader("{"))
	res := httptest.NewRecorder()

	appmw.JSONErrors(http.HandlerFunc(handler.Login)).ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
	if !strings.Contains(res.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("content type = %q, want JSON", res.Header().Get("Content-Type"))
	}
}
