package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrWaiterNotFound      = errors.New("waiter not found")
	ErrCustomerNotVerified = errors.New("customer is not verified")
	ErrWaiterInactive      = errors.New("waiter inactive")
	ErrOrderStatusConflict = errors.New("order status conflict")
	ErrOrderStatusInvalid  = errors.New("order status invalid")
)

type Waiter struct {
	ID         int64
	CustomerID int64
	Name       string
	Phone      string
	Active     bool
}

type WaiterSessionStore struct {
	db *pgxpool.Pool
}

func NewWaiterSessionStore(db *pgxpool.Pool) *WaiterSessionStore {
	return &WaiterSessionStore{db: db}
}

func (s *WaiterSessionStore) Provision(ctx context.Context, phone, name string) (Waiter, error) {
	var waiter Waiter
	err := s.db.QueryRow(ctx, `
		WITH verified_customer AS (
			SELECT id, phone_number
			FROM customers
			WHERE phone_number = $1 AND phone_verified = TRUE AND is_active = TRUE
		)
		INSERT INTO waiter_staff (customer_id, name)
		SELECT id, $2 FROM verified_customer
		ON CONFLICT (customer_id) DO UPDATE SET
			name = EXCLUDED.name,
			is_active = TRUE,
			updated_at = NOW()
		RETURNING id, customer_id, name, TRUE`, phone, name).Scan(
		&waiter.ID, &waiter.CustomerID, &waiter.Name, &waiter.Active)
	if errors.Is(err, pgx.ErrNoRows) {
		return Waiter{}, ErrCustomerNotVerified
	}
	if err != nil {
		return Waiter{}, fmt.Errorf("provision waiter: %w", err)
	}
	waiter.Phone = phone
	return waiter, nil
}

func (s *WaiterSessionStore) FindByPhone(ctx context.Context, phone string) (Waiter, error) {
	var waiter Waiter
	err := s.db.QueryRow(ctx, `
		SELECT ws.id, ws.customer_id, ws.name, c.phone_number, ws.is_active
		FROM waiter_staff ws
		INNER JOIN customers c ON c.id = ws.customer_id
		WHERE c.phone_number = $1 AND c.phone_verified = TRUE
	`, phone).Scan(&waiter.ID, &waiter.CustomerID, &waiter.Name, &waiter.Phone, &waiter.Active)
	if errors.Is(err, pgx.ErrNoRows) {
		return Waiter{}, ErrWaiterNotFound
	}
	if err != nil {
		return Waiter{}, fmt.Errorf("find waiter: %w", err)
	}
	if !waiter.Active {
		return Waiter{}, ErrWaiterInactive
	}
	return waiter, nil
}

func (s *WaiterSessionStore) CreateSession(ctx context.Context, waiterID int64, tokenHash []byte, expiresAt time.Time) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO waiter_sessions (waiter_id, session_token_hash, expires_at)
		VALUES ($1, $2, $3)`, waiterID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("create waiter session: %w", err)
	}
	return nil
}

func (s *WaiterSessionStore) Lookup(ctx context.Context, tokenHash []byte) (Waiter, bool, error) {
	var waiter Waiter
	err := s.db.QueryRow(ctx, `
		SELECT ws.id, ws.customer_id, ws.name, c.phone_number, ws.is_active
		FROM waiter_sessions ss
		INNER JOIN waiter_staff ws ON ws.id = ss.waiter_id
		INNER JOIN customers c ON c.id = ws.customer_id
		WHERE ss.session_token_hash = $1
			AND ss.expires_at > NOW()
			AND ss.revoked_at IS NULL
			AND ws.is_active = TRUE
			AND c.is_active = TRUE`, tokenHash).Scan(
		&waiter.ID, &waiter.CustomerID, &waiter.Name, &waiter.Phone, &waiter.Active)
	if errors.Is(err, pgx.ErrNoRows) {
		return Waiter{}, false, nil
	}
	if err != nil {
		return Waiter{}, false, fmt.Errorf("lookup waiter session: %w", err)
	}
	if _, err := s.db.Exec(ctx, `
		UPDATE waiter_sessions SET last_seen_at = NOW()
		WHERE session_token_hash = $1 AND revoked_at IS NULL`, tokenHash); err != nil {
		return Waiter{}, false, fmt.Errorf("update waiter session activity: %w", err)
	}
	return waiter, true, nil
}

func (s *WaiterSessionStore) Revoke(ctx context.Context, tokenHash []byte) error {
	if _, err := s.db.Exec(ctx, `
		UPDATE waiter_sessions SET revoked_at = COALESCE(revoked_at, NOW())
		WHERE session_token_hash = $1`, tokenHash); err != nil {
		return fmt.Errorf("revoke waiter session: %w", err)
	}
	return nil
}

func (s *WaiterSessionStore) ListOrders(ctx context.Context) ([]Order, error) {
	rows, err := s.db.Query(ctx, `
		SELECT o.id
		FROM orders o
		WHERE o.status IN ('new', 'accepted', 'preparing', 'ready', 'served')
		ORDER BY CASE o.status
			WHEN 'new' THEN 0
			WHEN 'accepted' THEN 1
			WHEN 'preparing' THEN 2
			WHEN 'ready' THEN 3
			WHEN 'served' THEN 4
		END, o.created_at ASC, o.id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list staff orders: %w", err)
	}
	defer rows.Close()
	var orders []Order
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan staff order id: %w", err)
		}
		order, err := s.getOrderByID(ctx, id)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate staff orders: %w", err)
	}
	return orders, nil
}

func (s *WaiterSessionStore) GetOrder(ctx context.Context, orderID int64) (Order, error) {
	return s.getOrderByID(ctx, orderID)
}

func (s *WaiterSessionStore) TransitionOrder(ctx context.Context, orderID int64, newStatus string, waiterID, adminID *int64, note string) (Order, error) {
	if !validStaffTransition("", newStatus) {
		return Order{}, ErrOrderStatusInvalid
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Order{}, fmt.Errorf("begin order transition: %w", err)
	}
	defer tx.Rollback(ctx)

	var current string
	if err := tx.QueryRow(ctx, `SELECT status FROM orders WHERE id = $1 FOR UPDATE`, orderID).Scan(&current); errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrOrderNotFound
	} else if err != nil {
		return Order{}, fmt.Errorf("lock order for transition: %w", err)
	}
	if !validStaffTransition(current, newStatus) {
		return Order{}, ErrOrderStatusConflict
	}

	_, err = tx.Exec(ctx, `
		UPDATE orders SET status = $2, updated_at = NOW(),
			accepted_at = CASE WHEN $2 = 'accepted' AND accepted_at IS NULL THEN NOW() ELSE accepted_at END,
			completed_at = CASE WHEN $2 IN ('completed', 'cancelled') THEN NOW() ELSE completed_at END
		WHERE id = $1`, orderID, newStatus)
	if err != nil {
		return Order{}, fmt.Errorf("update order status: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO order_status_history (
			order_id, from_status, to_status, previous_status, new_status,
			changed_by, waiter_id, changed_at, created_at, note
		) VALUES ($1, $2, $3, $2, $3, $4, $5, NOW(), NOW(), NULLIF($6, ''))`,
		orderID, current, newStatus, adminID, waiterID, note); err != nil {
		return Order{}, fmt.Errorf("record order transition: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Order{}, fmt.Errorf("commit order transition: %w", err)
	}
	return s.getOrderByID(ctx, orderID)
}

func (s *WaiterSessionStore) getOrderByID(ctx context.Context, orderID int64) (Order, error) {
	// Staff access is intentionally separate from customer ownership. The
	// staff response still uses the customer-safe order shape.
	var order Order
	err := s.db.QueryRow(ctx, `
		SELECT o.id, o.order_number, COALESCE(o.customer_id, 0), o.table_id,
			o.table_session_id, o.status, o.subtotal::double precision, o.currency,
			COALESCE(o.customer_note, ''), o.created_at, o.updated_at,
			COALESCE(rt.table_number, '')
		FROM orders o
		LEFT JOIN restaurant_tables rt ON rt.id = o.table_id
		WHERE o.id = $1`, orderID).Scan(
		&order.ID, &order.OrderNumber, &order.CustomerID, &order.TableID,
		&order.TableSessionID, &order.Status, &order.Subtotal, &order.Currency,
		&order.CustomerNote, &order.CreatedAt, &order.UpdatedAt, &order.TableNumber)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrOrderNotFound
	}
	if err != nil {
		return Order{}, fmt.Errorf("load staff order: %w", err)
	}
	items, err := s.db.Query(ctx, `
		SELECT id, menu_item_id, item_name_snapshot, selected_variant, quantity,
			unit_price_snapshot::double precision, line_total::double precision
		FROM order_items WHERE order_id = $1 ORDER BY id`, orderID)
	if err != nil {
		return Order{}, fmt.Errorf("load staff order items: %w", err)
	}
	defer items.Close()
	order.Items = make([]OrderItem, 0)
	for items.Next() {
		var item OrderItem
		if err := items.Scan(&item.ID, &item.MenuItemID, &item.ItemNameSnapshot, &item.SelectedVariant, &item.Quantity, &item.UnitPriceSnapshot, &item.LineTotal); err != nil {
			return Order{}, fmt.Errorf("scan staff order item: %w", err)
		}
		order.Items = append(order.Items, item)
	}
	history, err := s.db.Query(ctx, `
		SELECT COALESCE(previous_status, ''), new_status, changed_at, COALESCE(note, '')
		FROM order_status_history WHERE order_id = $1 ORDER BY changed_at, id`, orderID)
	if err != nil {
		return Order{}, fmt.Errorf("load staff order history: %w", err)
	}
	defer history.Close()
	order.StatusHistory = make([]OrderStatusHistory, 0)
	for history.Next() {
		var entry OrderStatusHistory
		if err := history.Scan(&entry.PreviousStatus, &entry.NewStatus, &entry.ChangedAt, &entry.Note); err != nil {
			return Order{}, fmt.Errorf("scan staff order history: %w", err)
		}
		order.StatusHistory = append(order.StatusHistory, entry)
	}
	return order, history.Err()
}

func validStaffTransition(current, next string) bool {
	if current == "" {
		return next == "accepted" || next == "cancelled" || next == "preparing" || next == "ready" || next == "served" || next == "completed"
	}
	switch current {
	case "new":
		return next == "accepted" || next == "cancelled"
	case "accepted":
		return next == "preparing" || next == "cancelled"
	case "preparing":
		return next == "ready" || next == "cancelled"
	case "ready":
		return next == "served"
	case "served":
		return next == "completed"
	default:
		return false
	}
}
