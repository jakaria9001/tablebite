package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	appmw "github.com/jakaria9001/tablebite/internal/middleware"
	"github.com/jakaria9001/tablebite/internal/repository"
)

type OrderCreator interface {
	Create(ctx context.Context, customerID int64, input repository.CreateOrderInput) (repository.Order, error)
}

type OrderHandler struct {
	store OrderCreator
}

func NewOrderHandler(store OrderCreator) *OrderHandler {
	return &OrderHandler{store: store}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	customer, ok := appmw.CustomerFromContext(r.Context())
	if !ok || !customer.IsActive || !customer.PhoneVerified {
		respondError(w, r, http.StatusUnauthorized, "customer_auth_required", "Customer authentication is required.")
		return
	}

	var payload struct {
		TableSessionID int64                       `json:"tableSessionId"`
		CustomerNote   string                      `json:"customerNote"`
		Items          []repository.OrderItemInput `json:"items"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		respondError(w, r, http.StatusBadRequest, "invalid_order", "Please provide a valid order.")
		return
	}
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if len(idempotencyKey) < 1 || len(idempotencyKey) > 128 {
		respondError(w, r, http.StatusBadRequest, "invalid_idempotency_key", "A valid idempotency key is required.")
		return
	}

	order, err := h.store.Create(r.Context(), customer.ID, repository.CreateOrderInput{
		TableSessionID: payload.TableSessionID,
		IdempotencyKey: idempotencyKey,
		CustomerNote:   strings.TrimSpace(payload.CustomerNote),
		Items:          payload.Items,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrOrderTableSessionInvalid):
			respondError(w, r, http.StatusConflict, "invalid_table_session", "This table session is no longer active.")
		case errors.Is(err, repository.ErrOrderMenuItemNotFound),
			errors.Is(err, repository.ErrOrderItemInactive):
			respondError(w, r, http.StatusBadRequest, "invalid_menu_item", "One or more selected items are no longer available.")
		case errors.Is(err, repository.ErrOrderItemUnavailable):
			respondError(w, r, http.StatusConflict, "item_unavailable", "One or more selected items are sold out.")
		case errors.Is(err, repository.ErrOrderVariantInvalid),
			errors.Is(err, repository.ErrOrderQuantityInvalid),
			errors.Is(err, repository.ErrOrderIdempotencyInvalid):
			respondError(w, r, http.StatusBadRequest, "invalid_order", "Please review the order details.")
		default:
			respondError(w, r, http.StatusServiceUnavailable, "order_unavailable", "We could not place the order. Please try again.")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"order": order})
}
