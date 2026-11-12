package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jakaria9001/tablebite/internal/repository"
)

type TableSessionStore interface {
	Resolve(ctx context.Context, qrToken string) (repository.TableSession, error)
	Get(ctx context.Context, sessionID int64) (repository.TableSession, error)
	Open(ctx context.Context, tableID int64) (repository.TableSession, error)
	Close(ctx context.Context, tableID int64) (repository.TableSession, error)
}

type TableSessionHandler struct {
	store TableSessionStore
}

func NewTableSessionHandler(store TableSessionStore) *TableSessionHandler {
	return &TableSessionHandler{store: store}
}

func (h *TableSessionHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "tableToken")
	session, err := h.store.Resolve(r.Context(), token)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeTableSession(w, session)
}

func (h *TableSessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	sessionID, err := strconv.ParseInt(chi.URLParam(r, "sessionId"), 10, 64)
	if err != nil || sessionID <= 0 {
		http.Error(w, "table session not found", http.StatusNotFound)
		return
	}
	session, err := h.store.Get(r.Context(), sessionID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeTableSession(w, session)
}

func (h *TableSessionHandler) Open(w http.ResponseWriter, r *http.Request) {
	tableID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || tableID <= 0 {
		http.Error(w, "invalid table id", http.StatusBadRequest)
		return
	}
	session, err := h.store.Open(r.Context(), tableID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeTableSession(w, session)
}

func (h *TableSessionHandler) Close(w http.ResponseWriter, r *http.Request) {
	tableID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || tableID <= 0 {
		http.Error(w, "invalid table id", http.StatusBadRequest)
		return
	}
	session, err := h.store.Close(r.Context(), tableID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeTableSession(w, session)
}

func (h *TableSessionHandler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrTableNotFound),
		errors.Is(err, repository.ErrTableSessionNotFound),
		errors.Is(err, repository.ErrTableInactive):
		http.Error(w, "table session not found", http.StatusNotFound)
	default:
		http.Error(w, "table session service unavailable", http.StatusServiceUnavailable)
	}
}

func writeTableSession(w http.ResponseWriter, session repository.TableSession) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"session": map[string]any{
			"id":          session.ID,
			"tableId":     session.TableID,
			"tableNumber": session.TableNumber,
			"tableName":   session.DisplayName,
			"qrToken":     session.QRToken,
			"openedAt":    session.OpenedAt,
			"closedAt":    session.ClosedAt,
			"active":      session.ClosedAt == nil,
		},
	})
}
