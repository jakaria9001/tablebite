package session

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ db *pgxpool.Pool }

func NewPostgresStore(db *pgxpool.Pool) *PostgresStore { return &PostgresStore{db: db} }

func (s *PostgresStore) Register(ctx context.Context, token string, user AdminUser) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO admin_sessions (token, admin_id, email, role, is_active, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (token) DO UPDATE SET
			admin_id = EXCLUDED.admin_id,
			email = EXCLUDED.email,
			role = EXCLUDED.role,
			is_active = EXCLUDED.is_active,
			expires_at = EXCLUDED.expires_at`,
		token, user.ID, user.Email, user.Role, user.Active, user.Expires)
	if err != nil {
		return fmt.Errorf("register session: %w", err)
	}
	return nil
}

func (s *PostgresStore) Lookup(ctx context.Context, token string) (AdminUser, bool, error) {
	var user AdminUser
	err := s.db.QueryRow(ctx, `
		SELECT admin_id, email, role, is_active, expires_at
		FROM admin_sessions
		WHERE token = $1 AND expires_at > NOW()`, token).
		Scan(&user.ID, &user.Email, &user.Role, &user.Active, &user.Expires)
	if errors.Is(err, pgx.ErrNoRows) {
		return AdminUser{}, false, nil
	}
	if err != nil {
		return AdminUser{}, false, fmt.Errorf("lookup session: %w", err)
	}
	return user, true, nil
}

func (s *PostgresStore) Clear(ctx context.Context, token string) error {
	if _, err := s.db.Exec(ctx, `DELETE FROM admin_sessions WHERE token = $1`, token); err != nil {
		return fmt.Errorf("clear session: %w", err)
	}
	return nil
}
