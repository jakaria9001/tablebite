package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	appmw "github.com/jakaria9001/tablebite/internal/middleware"
	"github.com/jakaria9001/tablebite/internal/repository"
)

type fakeOrderCreator struct {
	mu       sync.Mutex
	inputs   []repository.CreateOrderInput
	order    repository.Order
	err      error
	sequence int
}

type fakeOrderCustomerStore struct {
	customer repository.Customer
}

type fakeOrderReader struct {
	order repository.Order
	err   error
}

func (f fakeOrderReader) Get(context.Context, int64, int64) (repository.Order, error) {
	return f.order, f.err
}

func (s fakeOrderCustomerStore) Lookup(context.Context, []byte) (repository.Customer, bool, error) {
	return s.customer, true, nil
}

func (fakeOrderCustomerStore) Revoke(context.Context, []byte) error {
	return nil
}

func (f *fakeOrderCreator) Create(_ context.Context, _ int64, input repository.CreateOrderInput) (repository.Order, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.inputs = append(f.inputs, input)
	f.sequence++
	if f.err != nil {
		return repository.Order{}, f.err
	}
	return f.order, nil
}

func TestCreateOrderRequiresAuthenticatedCustomer(t *testing.T) {
	handler := NewOrderHandler(&fakeOrderCreator{})
	req := orderRequest(`{"tableSessionId":42,"items":[{"menuItemId":7,"variant":"regular","quantity":1}]}`, "key-1")
	res := httptest.NewRecorder()

	handler.Create(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
}

func TestCreateOrderNormalAndIdempotentRequest(t *testing.T) {
	store := &fakeOrderCreator{order: testOrder()}
	handler := NewOrderHandler(store)
	body := `{"tableSessionId":42,"items":[{"menuItemId":7,"variant":"regular","quantity":2}]}`

	for i := 0; i < 2; i++ {
		req := orderRequest(body, "same-request")
		res := httptest.NewRecorder()
		authed := appmw.WithCustomerUser(fakeOrderCustomerStore{customer: repository.Customer{ID: 9, PhoneVerified: true, IsActive: true}})(http.HandlerFunc(handler.Create))
		authed.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, want %d", i, res.Code, http.StatusOK)
		}
		var response struct {
			Order repository.Order `json:"order"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.Order.Status != "new" {
			t.Fatalf("status = %q, want new", response.Order.Status)
		}
	}
	if len(store.inputs) != 2 || store.inputs[0].IdempotencyKey != "same-request" {
		t.Fatalf("unexpected store inputs: %#v", store.inputs)
	}
}

func TestCreateOrderRejectsPriceTampering(t *testing.T) {
	handler := NewOrderHandler(&fakeOrderCreator{})
	req := orderRequest(`{"tableSessionId":42,"items":[{"menuItemId":7,"variant":"regular","quantity":1,"unitPrice":1}]}`, "key-1")
	res := httptest.NewRecorder()

	authed := appmw.WithCustomerUser(fakeOrderCustomerStore{customer: repository.Customer{ID: 9, PhoneVerified: true, IsActive: true}})(http.HandlerFunc(handler.Create))
	authed.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
}

func TestCreateOrderMapsValidationFailures(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{"invalid table session", repository.ErrOrderTableSessionInvalid, http.StatusConflict},
		{"sold out item", repository.ErrOrderItemUnavailable, http.StatusConflict},
		{"invalid variant", repository.ErrOrderVariantInvalid, http.StatusBadRequest},
		{"database failure", errors.New("database unavailable"), http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewOrderHandler(&fakeOrderCreator{err: tt.err})
			req := orderRequest(`{"tableSessionId":42,"items":[{"menuItemId":7,"variant":"half","quantity":1}]}`, "key-1")
			res := httptest.NewRecorder()

			authed := appmw.WithCustomerUser(fakeOrderCustomerStore{customer: repository.Customer{ID: 9, PhoneVerified: true, IsActive: true}})(http.HandlerFunc(handler.Create))
			authed.ServeHTTP(res, req)

			if res.Code != tt.status {
				t.Fatalf("status = %d, want %d", res.Code, tt.status)
			}
		})
	}
}

func TestCreateOrderConcurrentSubmissionsUseIdempotencyKey(t *testing.T) {
	store := &fakeOrderCreator{order: testOrder()}
	handler := NewOrderHandler(store)
	const submissions = 8
	var wg sync.WaitGroup
	responses := make(chan int, submissions)

	for i := 0; i < submissions; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := orderRequest(`{"tableSessionId":42,"items":[{"menuItemId":7,"variant":"full","quantity":1}]}`, "concurrent-key")
			res := httptest.NewRecorder()
			authed := appmw.WithCustomerUser(fakeOrderCustomerStore{customer: repository.Customer{ID: 9, PhoneVerified: true, IsActive: true}})(http.HandlerFunc(handler.Create))
			authed.ServeHTTP(res, req)
			responses <- res.Code
		}()
	}

	wg.Wait()
	close(responses)

	for status := range responses {
		if status != http.StatusOK {
			t.Fatalf("concurrent status = %d, want %d", status, http.StatusOK)
		}
	}

	if len(store.inputs) != submissions {
		t.Fatalf("store calls = %d, want %d", len(store.inputs), submissions)
	}
}

func TestOrderDetailsRequiresOwnershipAndAuthentication(t *testing.T) {
	reader := fakeOrderReader{order: testOrder()}
	handler := NewOrderDetailsHandler(reader)

	unauthenticated := httptest.NewRequest(http.MethodGet, "/api/v1/orders/100", nil)
	unauthenticated = withChiParams(unauthenticated, map[string]string{"orderId": "100"})
	unauthenticatedResponse := httptest.NewRecorder()
	handler.Get(unauthenticatedResponse, unauthenticated)
	if unauthenticatedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want %d", unauthenticatedResponse.Code, http.StatusUnauthorized)
	}

	owned := httptest.NewRequest(http.MethodGet, "/api/v1/orders/100", nil)
	owned = withChiParams(owned, map[string]string{"orderId": "100"})
	owned.AddCookie(&http.Cookie{Name: "tablebite_customer_session", Value: "test-session"})
	ownedResponse := httptest.NewRecorder()
	authed := appmw.WithCustomerUser(fakeOrderCustomerStore{customer: repository.Customer{ID: 9, PhoneVerified: true, IsActive: true}})(http.HandlerFunc(handler.Get))
	authed.ServeHTTP(ownedResponse, owned)
	if ownedResponse.Code != http.StatusOK {
		t.Fatalf("owned order status = %d, want %d", ownedResponse.Code, http.StatusOK)
	}
	if !strings.Contains(ownedResponse.Body.String(), `"statusHistory"`) {
		t.Fatalf("response does not include status history: %s", ownedResponse.Body.String())
	}

	otherCustomer := NewOrderDetailsHandler(fakeOrderReader{err: repository.ErrOrderNotFound})
	otherRequest := httptest.NewRequest(http.MethodGet, "/api/v1/orders/100", nil)
	otherRequest = withChiParams(otherRequest, map[string]string{"orderId": "100"})
	otherRequest.AddCookie(&http.Cookie{Name: "tablebite_customer_session", Value: "test-session"})
	otherResponse := httptest.NewRecorder()
	appmw.WithCustomerUser(fakeOrderCustomerStore{customer: repository.Customer{ID: 11, PhoneVerified: true, IsActive: true}})(http.HandlerFunc(otherCustomer.Get)).ServeHTTP(otherResponse, otherRequest)
	if otherResponse.Code != http.StatusNotFound {
		t.Fatalf("other customer status = %d, want %d", otherResponse.Code, http.StatusNotFound)
	}

	invalidID := httptest.NewRequest(http.MethodGet, "/api/v1/orders/not-an-id", nil)
	invalidID = withChiParams(invalidID, map[string]string{"orderId": "not-an-id"})
	invalidID.AddCookie(&http.Cookie{Name: "tablebite_customer_session", Value: "test-session"})
	invalidResponse := httptest.NewRecorder()
	authed.ServeHTTP(invalidResponse, invalidID)
	if invalidResponse.Code != http.StatusNotFound {
		t.Fatalf("invalid order id status = %d, want %d", invalidResponse.Code, http.StatusNotFound)
	}
}
func orderRequest(body, idempotencyKey string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", idempotencyKey)
	req.AddCookie(&http.Cookie{Name: "tablebite_customer_session", Value: "test-session"})
	return req
}

func testOrder() repository.Order {
	return repository.Order{
		ID:             100,
		OrderNumber:    "TB-20261031-000100",
		CustomerID:     9,
		TableID:        7,
		TableSessionID: 42,
		Status:         "new",
		Subtotal:       240,
		Currency:       "INR",
		StatusHistory: []repository.OrderStatusHistory{{
			NewStatus: "new",
		}},
	}
}
