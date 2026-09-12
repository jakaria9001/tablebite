package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jakaria9001/tablebite/internal/model"
)

var ErrAdminNotFound = errors.New("admin not found")

type AdminRepository struct{ db *pgxpool.Pool }

func NewAdminRepository(db *pgxpool.Pool) *AdminRepository { return &AdminRepository{db: db} }

func (r *AdminRepository) FindByEmail(ctx context.Context, email string) (model.Admin, error) {
	var admin model.Admin
	err := r.db.QueryRow(ctx, `SELECT id, email, password_hash, role, is_active FROM admins WHERE email = $1`, email).
		Scan(&admin.ID, &admin.Email, &admin.PasswordHash, &admin.Role, &admin.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return admin, ErrAdminNotFound
	}
	if err != nil {
		return admin, fmt.Errorf("load admin: %w", err)
	}
	return admin, nil
}

func (r *AdminRepository) UpdatePasswordHash(ctx context.Context, adminID int64, hash string) error {
	_, err := r.db.Exec(ctx, `UPDATE admins SET password_hash = $2, updated_at = NOW() WHERE id = $1`, adminID, hash)
	if err != nil {
		return fmt.Errorf("update admin password: %w", err)
	}
	return nil
}
