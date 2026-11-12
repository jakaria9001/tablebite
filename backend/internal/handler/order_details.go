package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	appmw "github.com/jakaria9001/tablebite/internal/middleware"
	"github.com/jakaria9001/tablebite/internal/repository"
)

type OrderReader interface {
	Get(ctx context.Context, customerID, orderID int64) (repository.Order, error)
}

type OrderDetailsHandler struct {
	store OrderReader
}

func NewOrderDetailsHandler(store OrderReader) *OrderDetailsHandler {
	return &OrderDetailsHandler{store: store}
}

func (h *OrderDetailsHandler) Get(w http.ResponseWriter, r *http.Request) {
	customer, ok := appmw.CustomerFromContext(r.Context())
	if !ok || !customer.IsActive || !customer.PhoneVerified {
		respondError(w, r, http.StatusUnauthorized, "customer_auth_required", "Customer authentication is required.")
		return
	}

	orderID, err := strconv.ParseInt(chi.URLParam(r, "orderId"), 10, 64)
	if err != nil || orderID <= 0 {
		respondError(w, r, http.StatusNotFound, "order_not_found", "Order not found.")
		return
	}

	order, err := h.store.Get(r.Context(), customer.ID, orderID)
	if errors.Is(err, repository.ErrOrderNotFound) {
		respondError(w, r, http.StatusNotFound, "order_not_found", "Order not found.")
		return
	}
	if err != nil {
		respondError(w, r, http.StatusServiceUnavailable, "order_unavailable", "We could not load the order. Please try again.")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"order": order})
}
