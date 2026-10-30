package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerSessionStore struct {
	db *pgxpool.Pool
}

func NewCustomerSessionStore(db *pgxpool.Pool) *CustomerSessionStore {
	return &CustomerSessionStore{db: db}
}

func (s *CustomerSessionStore) Lookup(ctx context.Context, tokenHash []byte) (Customer, bool, error) {
	var customer Customer
	err := s.db.QueryRow(ctx, `
		SELECT c.id, c.phone_number, c.phone_verified, c.name, c.is_active, c.last_login_at
		FROM customer_sessions cs
		INNER JOIN customers c ON c.id = cs.customer_id
		WHERE cs.session_token_hash = $1
			AND cs.expires_at > NOW()
			AND cs.revoked_at IS NULL
			AND c.is_active = TRUE`, tokenHash,
	).Scan(
		&customer.ID,
		&customer.PhoneNumber,
		&customer.PhoneVerified,
		&customer.Name,
		&customer.IsActive,
		&customer.LastLoginAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Customer{}, false, nil
	}
	if err != nil {
		return Customer{}, false, fmt.Errorf("lookup customer session: %w", err)
	}
	if _, err := s.db.Exec(ctx, `
		UPDATE customer_sessions
		SET last_seen_at = NOW()
		WHERE session_token_hash = $1 AND revoked_at IS NULL`, tokenHash); err != nil {
		return Customer{}, false, fmt.Errorf("update customer session activity: %w", err)
	}
	return customer, true, nil
}

func (s *CustomerSessionStore) Revoke(ctx context.Context, tokenHash []byte) error {
	if _, err := s.db.Exec(ctx, `
		UPDATE customer_sessions
		SET revoked_at = COALESCE(revoked_at, NOW())
		WHERE session_token_hash = $1`, tokenHash); err != nil {
		return fmt.Errorf("revoke customer session: %w", err)
	}
	return nil
}
