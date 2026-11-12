package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jakaria9001/tablebite/internal/repository"
)

type fakeTableSessionStore struct {
	session repository.TableSession
	err     error
	opens   int64
	closes  int64
}

func (s *fakeTableSessionStore) Resolve(context.Context, string) (repository.TableSession, error) {
	return s.session, s.err
}

func (s *fakeTableSessionStore) Get(context.Context, int64) (repository.TableSession, error) {
	return s.session, s.err
}

func (s *fakeTableSessionStore) Open(_ context.Context, tableID int64) (repository.TableSession, error) {
	s.opens = tableID
	return s.session, s.err
}

func (s *fakeTableSessionStore) Close(_ context.Context, tableID int64) (repository.TableSession, error) {
	s.closes = tableID
	return s.session, s.err
}

func TestResolveTableSessionReturnsValidatedSession(t *testing.T) {
	store := &fakeTableSessionStore{session: testTableSession()}
	handler := NewTableSessionHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tables/resolve/qr-token", nil)
	req = withChiParams(req, map[string]string{"tableToken": "qr-token"})
	res := httptest.NewRecorder()

	handler.Resolve(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if !strings.Contains(res.Body.String(), `"id":42`) || !strings.Contains(res.Body.String(), `"tableNumber":"07"`) {
		t.Fatalf("response = %s", res.Body.String())
	}
}

func TestGetClosedSessionReturnsNotFound(t *testing.T) {
	store := &fakeTableSessionStore{err: repository.ErrTableSessionNotFound}
	handler := NewTableSessionHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tables/session/42", nil)
	req = withChiParams(req, map[string]string{"sessionId": "42"})
	res := httptest.NewRecorder()

	handler.Get(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNotFound)
	}
}

func TestAdminTableSessionLifecycleUsesTableID(t *testing.T) {
	store := &fakeTableSessionStore{session: testTableSession()}
	handler := NewTableSessionHandler(store)

	openReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/tables/7/sessions/open", nil)
	openReq = withChiParams(openReq, map[string]string{"id": "7"})
	openRes := httptest.NewRecorder()
	handler.Open(openRes, openReq)

	closeReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/tables/7/sessions/close", nil)
	closeReq = withChiParams(closeReq, map[string]string{"id": "7"})
	closeRes := httptest.NewRecorder()
	handler.Close(closeRes, closeReq)

	if openRes.Code != http.StatusOK || closeRes.Code != http.StatusOK {
		t.Fatalf("open/close statuses = %d/%d", openRes.Code, closeRes.Code)
	}
	if store.opens != 7 || store.closes != 7 {
		t.Fatalf("table ids = open %d, close %d", store.opens, store.closes)
	}
}

func testTableSession() repository.TableSession {
	return repository.TableSession{
		ID:          42,
		TableID:     7,
		TableNumber: "07",
		DisplayName: "Table 07",
		QRToken:     "qr-token",
		OpenedAt:    time.Now(),
	}
}

func withChiParams(req *http.Request, params map[string]string) *http.Request {
	ctx := chi.NewRouteContext()
	for key, value := range params {
		ctx.URLParams.Add(key, value)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
}
