package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	MaxOrderItemQuantity = 20
	MaxOrderItems        = 50
)

var (
	ErrOrderTableSessionInvalid = errors.New("order table session invalid")
	ErrOrderMenuItemNotFound    = errors.New("order menu item not found")
	ErrOrderItemUnavailable     = errors.New("order menu item unavailable")
	ErrOrderItemInactive        = errors.New("order menu item inactive")
	ErrOrderVariantInvalid      = errors.New("order variant invalid")
	ErrOrderQuantityInvalid     = errors.New("order quantity invalid")
	ErrOrderIdempotencyInvalid  = errors.New("order idempotency key invalid")
	ErrOrderNotFound            = errors.New("order not found")
)

type OrderItemInput struct {
	MenuItemID int64  `json:"menuItemId"`
	Variant    string `json:"variant"`
	Quantity   int    `json:"quantity"`
}

type CreateOrderInput struct {
	TableSessionID int64
	IdempotencyKey string
	CustomerNote   string
	Items          []OrderItemInput
}

type Order struct {
	ID             int64                `json:"id"`
	OrderNumber    string               `json:"orderNumber"`
	CustomerID     int64                `json:"customerId"`
	TableID        int64                `json:"tableId"`
	TableSessionID int64                `json:"tableSessionId"`
	Status         string               `json:"status"`
	Subtotal       float64              `json:"subtotal"`
	Currency       string               `json:"currency"`
	CustomerNote   string               `json:"customerNote,omitempty"`
	CreatedAt      time.Time            `json:"createdAt"`
	UpdatedAt      time.Time            `json:"updatedAt"`
	Items          []OrderItem          `json:"items"`
	StatusHistory  []OrderStatusHistory `json:"statusHistory"`
	TableNumber    string               `json:"tableNumber,omitempty"`
}

type OrderItem struct {
	ID                int64   `json:"id"`
	MenuItemID        int64   `json:"menuItemId"`
	ItemNameSnapshot  string  `json:"itemName"`
	SelectedVariant   string  `json:"variant"`
	Quantity          int     `json:"quantity"`
	UnitPriceSnapshot float64 `json:"unitPrice"`
	LineTotal         float64 `json:"lineTotal"`
}

type OrderStatusHistory struct {
	PreviousStatus string    `json:"previousStatus,omitempty"`
	NewStatus      string    `json:"newStatus"`
	ChangedAt      time.Time `json:"changedAt"`
	Note           string    `json:"note,omitempty"`
}

type OrderStore struct {
	db *pgxpool.Pool
}

func NewOrderStore(db *pgxpool.Pool) *OrderStore {
	return &OrderStore{db: db}
}

func (s *OrderStore) Get(ctx context.Context, customerID, orderID int64) (Order, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Order{}, fmt.Errorf("begin order lookup: %w", err)
	}
	defer tx.Rollback(ctx)

	var ownedID int64
	if err := tx.QueryRow(ctx, `
		SELECT id
		FROM orders
		WHERE id = $1 AND customer_id = $2`, orderID, customerID).Scan(&ownedID); errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrOrderNotFound
	} else if err != nil {
		return Order{}, fmt.Errorf("check order ownership: %w", err)
	}

	order, err := loadOrder(ctx, tx, ownedID)
	if err != nil {
		return Order{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Order{}, fmt.Errorf("commit order lookup: %w", err)
	}
	return order, nil
}

func (s *OrderStore) Create(ctx context.Context, customerID int64, input CreateOrderInput) (Order, error) {
	if customerID <= 0 || input.TableSessionID <= 0 || len(input.Items) == 0 || len(input.Items) > MaxOrderItems {
		return Order{}, ErrOrderQuantityInvalid
	}
	if len(input.IdempotencyKey) < 1 || len(input.IdempotencyKey) > 128 {
		return Order{}, ErrOrderIdempotencyInvalid
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Order{}, fmt.Errorf("begin order transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var restaurantID, tableID int64
	var restaurantCurrency string
	err = tx.QueryRow(ctx, `
		SELECT rt.restaurant_id, ts.table_id, r.currency
		FROM table_sessions ts
		INNER JOIN restaurant_tables rt ON rt.id = ts.table_id
		INNER JOIN restaurants r ON r.id = rt.restaurant_id
		WHERE ts.id = $1
			AND ts.closed_at IS NULL
			AND rt.is_active = TRUE
			AND r.is_active = TRUE
		FOR SHARE OF ts, rt, r`, input.TableSessionID).Scan(&restaurantID, &tableID, &restaurantCurrency)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrOrderTableSessionInvalid
	}
	if err != nil {
		return Order{}, fmt.Errorf("validate table session: %w", err)
	}

	var existingID int64
	err = tx.QueryRow(ctx, `
		SELECT id
		FROM orders
		WHERE restaurant_id = $1 AND customer_id = $2 AND idempotency_key = $3
		LIMIT 1`, restaurantID, customerID, input.IdempotencyKey).Scan(&existingID)
	if err == nil {
		order, loadErr := loadOrder(ctx, tx, existingID)
		if loadErr != nil {
			return Order{}, loadErr
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return Order{}, fmt.Errorf("commit duplicate order lookup: %w", commitErr)
		}
		return order, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Order{}, fmt.Errorf("check order idempotency: %w", err)
	}

	type menuItem struct {
		ID          int64
		Name        string
		Price       *float64
		PriceHalf   *float64
		PriceFull   *float64
		Currency    string
		IsAvailable bool
		IsActive    bool
	}
	menuItems := make(map[int64]menuItem, len(input.Items))
	for _, item := range input.Items {
		if item.MenuItemID <= 0 || item.Quantity < 1 || item.Quantity > MaxOrderItemQuantity {
			return Order{}, ErrOrderQuantityInvalid
		}
		if item.Variant != "regular" && item.Variant != "half" && item.Variant != "full" {
			return Order{}, ErrOrderVariantInvalid
		}
		if _, exists := menuItems[item.MenuItemID]; exists {
			return Order{}, ErrOrderQuantityInvalid
		}

		var current menuItem
		err := tx.QueryRow(ctx, `
			SELECT id, name, price::double precision, price_half::double precision,
				price_full::double precision, currency, is_available, is_active
			FROM menu_items
			WHERE id = $1 AND restaurant_id = $2
			FOR SHARE`, item.MenuItemID, restaurantID).Scan(
			&current.ID, &current.Name, &current.Price, &current.PriceHalf, &current.PriceFull,
			&current.Currency, &current.IsAvailable, &current.IsActive)
		if errors.Is(err, pgx.ErrNoRows) {
			return Order{}, ErrOrderMenuItemNotFound
		}
		if err != nil {
			return Order{}, fmt.Errorf("load menu item: %w", err)
		}
		if !current.IsActive {
			return Order{}, ErrOrderItemInactive
		}
		if !current.IsAvailable {
			return Order{}, ErrOrderItemUnavailable
		}
		if priceForVariant(current.Price, current.PriceHalf, current.PriceFull, item.Variant) == nil {
			return Order{}, ErrOrderVariantInvalid
		}
		menuItems[item.MenuItemID] = current
	}

	var subtotal float64
	for _, item := range input.Items {
		price := priceForVariant(menuItems[item.MenuItemID].Price, menuItems[item.MenuItemID].PriceHalf, menuItems[item.MenuItemID].PriceFull, item.Variant)
		subtotal += *price * float64(item.Quantity)
	}
	subtotal = money(subtotal)
	if restaurantCurrency == "" {
		restaurantCurrency = "INR"
	}

	var orderID int64
	if err := tx.QueryRow(ctx, `SELECT nextval(pg_get_serial_sequence('orders', 'id'))`).Scan(&orderID); err != nil {
		return Order{}, fmt.Errorf("reserve order id: %w", err)
	}
	orderNumber := fmt.Sprintf("TB-%s-%06d", time.Now().UTC().Format("20060102"), orderID)

	var insertedID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (
			id, restaurant_id, customer_id, table_id, table_session_id, order_number,
			status, customer_note, subtotal, total, currency, idempotency_key
		) VALUES ($1, $2, $3, $4, $5, $6, 'new', NULLIF($7, ''), $8, $8, $9, $10)
		ON CONFLICT (restaurant_id, customer_id, idempotency_key)
			WHERE idempotency_key IS NOT NULL
		DO NOTHING
		RETURNING id`,
		orderID, restaurantID, customerID, tableID, input.TableSessionID, orderNumber,
		input.CustomerNote, subtotal, restaurantCurrency, input.IdempotencyKey).Scan(&insertedID)
	if errors.Is(err, pgx.ErrNoRows) {
		if err := tx.QueryRow(ctx, `
			SELECT id
			FROM orders
			WHERE restaurant_id = $1 AND customer_id = $2 AND idempotency_key = $3`,
			restaurantID, customerID, input.IdempotencyKey).Scan(&insertedID); err != nil {
			return Order{}, fmt.Errorf("load concurrent idempotent order: %w", err)
		}
		order, loadErr := loadOrder(ctx, tx, insertedID)
		if loadErr != nil {
			return Order{}, loadErr
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return Order{}, fmt.Errorf("commit concurrent order lookup: %w", commitErr)
		}
		return order, nil
	}
	if err != nil {
		return Order{}, fmt.Errorf("create order: %w", err)
	}

	for _, item := range input.Items {
		menuItem := menuItems[item.MenuItemID]
		price := *priceForVariant(menuItem.Price, menuItem.PriceHalf, menuItem.PriceFull, item.Variant)
		lineTotal := money(price * float64(item.Quantity))
		if _, err := tx.Exec(ctx, `
			INSERT INTO order_items (
				order_id, menu_item_id, item_name, item_name_snapshot, selected_variant,
				unit_price, unit_price_snapshot, quantity, line_total, created_at
			) VALUES ($1, $2, $3, $3, $4, $5, $5, $6, $7, NOW())`,
			insertedID, item.MenuItemID, menuItem.Name, item.Variant, price, item.Quantity, lineTotal); err != nil {
			return Order{}, fmt.Errorf("create order item: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO order_status_history (
			order_id, from_status, to_status, previous_status, new_status,
			changed_at, created_at, note
		) VALUES ($1, NULL, 'new', NULL, 'new', NOW(), NOW(), 'Order placed')`, insertedID); err != nil {
		return Order{}, fmt.Errorf("create order status history: %w", err)
	}

	order, err := loadOrder(ctx, tx, insertedID)
	if err != nil {
		return Order{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Order{}, fmt.Errorf("commit order transaction: %w", err)
	}
	return order, nil
}

func loadOrder(ctx context.Context, tx pgx.Tx, orderID int64) (Order, error) {
	order := Order{
		Items:         make([]OrderItem, 0),
		StatusHistory: make([]OrderStatusHistory, 0),
	}
	err := tx.QueryRow(ctx, `
		SELECT o.id, o.order_number, o.customer_id, o.table_id, o.table_session_id, o.status,
			o.subtotal::double precision, o.currency, COALESCE(o.customer_note, ''), o.created_at, o.updated_at,
			COALESCE(rt.table_number, '')
		FROM orders o
		LEFT JOIN restaurant_tables rt ON rt.id = o.table_id
		WHERE o.id = $1`, orderID).Scan(
		&order.ID, &order.OrderNumber, &order.CustomerID, &order.TableID, &order.TableSessionID,
		&order.Status, &order.Subtotal, &order.Currency, &order.CustomerNote, &order.CreatedAt, &order.UpdatedAt, &order.TableNumber)
	if err != nil {
		return Order{}, fmt.Errorf("load created order: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id, menu_item_id, item_name_snapshot, selected_variant, quantity,
			unit_price_snapshot::double precision, line_total::double precision
		FROM order_items
		WHERE order_id = $1
		ORDER BY id`, orderID)
	if err != nil {
		return Order{}, fmt.Errorf("load order items: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ID, &item.MenuItemID, &item.ItemNameSnapshot, &item.SelectedVariant,
			&item.Quantity, &item.UnitPriceSnapshot, &item.LineTotal); err != nil {
			return Order{}, fmt.Errorf("scan order item: %w", err)
		}
		order.Items = append(order.Items, item)
	}
	if err := rows.Err(); err != nil {
		return Order{}, fmt.Errorf("iterate order items: %w", err)
	}

	historyRows, err := tx.Query(ctx, `
		SELECT COALESCE(previous_status, ''), new_status, changed_at, COALESCE(note, '')
		FROM order_status_history
		WHERE order_id = $1
		ORDER BY changed_at, id`, orderID)
	if err != nil {
		return Order{}, fmt.Errorf("load order status history: %w", err)
	}
	defer historyRows.Close()
	for historyRows.Next() {
		var history OrderStatusHistory
		if err := historyRows.Scan(&history.PreviousStatus, &history.NewStatus, &history.ChangedAt, &history.Note); err != nil {
			return Order{}, fmt.Errorf("scan order status history: %w", err)
		}
		order.StatusHistory = append(order.StatusHistory, history)
	}
	if err := historyRows.Err(); err != nil {
		return Order{}, fmt.Errorf("iterate order status history: %w", err)
	}
	return order, nil
}

func priceForVariant(regular, half, full *float64, variant string) *float64 {
	switch variant {
	case "regular":
		return regular
	case "half":
		return half
	case "full":
		return full
	default:
		return nil
	}
}

func money(value float64) float64 {
	return math.Round(value*100) / 100
}
