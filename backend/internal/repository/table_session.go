package repository

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTableNotFound = errors.New("table not found")
var ErrTableInactive = errors.New("table inactive")
var ErrTableSessionNotFound = errors.New("table session not found")

type TableSession struct {
	ID          int64
	TableID     int64
	TableNumber string
	DisplayName string
	QRToken     string
	OpenedAt    time.Time
	ClosedAt    *time.Time
}

type TableSessionStore struct {
	db *pgxpool.Pool
}

func NewTableSessionStore(db *pgxpool.Pool) *TableSessionStore {
	return &TableSessionStore{db: db}
}

func (s *TableSessionStore) Resolve(ctx context.Context, qrToken string) (TableSession, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return TableSession{}, fmt.Errorf("begin table session resolution: %w", err)
	}
	defer tx.Rollback(ctx)

	var tableID int64
	err = tx.QueryRow(ctx, `
		SELECT rt.id
		FROM restaurant_tables rt
		INNER JOIN restaurants r ON r.id = rt.restaurant_id
		WHERE rt.qr_token = $1 AND rt.is_active = TRUE AND r.is_active = TRUE
		FOR UPDATE`, qrToken).Scan(&tableID)
	if errors.Is(err, pgx.ErrNoRows) {
		return TableSession{}, ErrTableNotFound
	}
	if err != nil {
		return TableSession{}, fmt.Errorf("resolve table QR token: %w", err)
	}

	session, err := openLockedTableSession(ctx, tx, tableID)
	if err != nil {
		return TableSession{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return TableSession{}, fmt.Errorf("commit table session resolution: %w", err)
	}
	return session, nil
}

func (s *TableSessionStore) Get(ctx context.Context, sessionID int64) (TableSession, error) {
	var session TableSession
	err := s.db.QueryRow(ctx, `
		SELECT ts.id, ts.table_id, rt.table_number, COALESCE(rt.display_name, ''),
			rt.qr_token, ts.opened_at, ts.closed_at
		FROM table_sessions ts
		INNER JOIN restaurant_tables rt ON rt.id = ts.table_id
		INNER JOIN restaurants r ON r.id = rt.restaurant_id
		WHERE ts.id = $1
			AND ts.closed_at IS NULL
			AND rt.is_active = TRUE
			AND r.is_active = TRUE`, sessionID).Scan(
		&session.ID,
		&session.TableID,
		&session.TableNumber,
		&session.DisplayName,
		&session.QRToken,
		&session.OpenedAt,
		&session.ClosedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return TableSession{}, ErrTableSessionNotFound
	}
	if err != nil {
		return TableSession{}, fmt.Errorf("load table session: %w", err)
	}
	return session, nil
}

func (s *TableSessionStore) Open(ctx context.Context, tableID int64) (TableSession, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return TableSession{}, fmt.Errorf("begin table session open: %w", err)
	}
	defer tx.Rollback(ctx)

	var active bool
	err = tx.QueryRow(ctx, `
		SELECT is_active
		FROM restaurant_tables rt
		INNER JOIN restaurants r ON r.id = rt.restaurant_id
		WHERE rt.id = $1 AND r.is_active = TRUE
		FOR UPDATE OF rt`, tableID).Scan(&active)
	if errors.Is(err, pgx.ErrNoRows) {
		return TableSession{}, ErrTableNotFound
	}
	if err != nil {
		return TableSession{}, fmt.Errorf("load table for session open: %w", err)
	}
	if !active {
		return TableSession{}, ErrTableInactive
	}

	session, err := openLockedTableSession(ctx, tx, tableID)
	if err != nil {
		return TableSession{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return TableSession{}, fmt.Errorf("commit table session open: %w", err)
	}
	return session, nil
}

func (s *TableSessionStore) Close(ctx context.Context, tableID int64) (TableSession, error) {
	var session TableSession
	err := s.db.QueryRow(ctx, `
		UPDATE table_sessions ts
		SET closed_at = NOW()
		FROM restaurant_tables rt
		INNER JOIN restaurants r ON r.id = rt.restaurant_id
		WHERE ts.table_id = rt.id
			AND rt.id = $1
			AND r.is_active = TRUE
			AND ts.closed_at IS NULL
		RETURNING ts.id, ts.table_id, rt.table_number, COALESCE(rt.display_name, ''),
			rt.qr_token, ts.opened_at, ts.closed_at`, tableID).Scan(
		&session.ID,
		&session.TableID,
		&session.TableNumber,
		&session.DisplayName,
		&session.QRToken,
		&session.OpenedAt,
		&session.ClosedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return TableSession{}, ErrTableSessionNotFound
	}
	if err != nil {
		return TableSession{}, fmt.Errorf("close table session: %w", err)
	}
	return session, nil
}

func openLockedTableSession(ctx context.Context, tx pgx.Tx, tableID int64) (TableSession, error) {
	var session TableSession
	err := tx.QueryRow(ctx, `
		SELECT ts.id, ts.table_id, rt.table_number, COALESCE(rt.display_name, ''),
			rt.qr_token, ts.opened_at, ts.closed_at
		FROM table_sessions ts
		INNER JOIN restaurant_tables rt ON rt.id = ts.table_id
		WHERE ts.table_id = $1 AND ts.closed_at IS NULL
		ORDER BY ts.opened_at DESC, ts.id DESC
		LIMIT 1`, tableID).Scan(
		&session.ID,
		&session.TableID,
		&session.TableNumber,
		&session.DisplayName,
		&session.QRToken,
		&session.OpenedAt,
		&session.ClosedAt,
	)
	if err == nil {
		return session, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return TableSession{}, fmt.Errorf("load active table session: %w", err)
	}

	token, err := generateTableSessionToken()
	if err != nil {
		return TableSession{}, fmt.Errorf("generate table session token: %w", err)
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO table_sessions (table_id, session_token)
		VALUES ($1, $2)
		RETURNING id, opened_at, closed_at`, tableID, token).Scan(
		&session.ID,
		&session.OpenedAt,
		&session.ClosedAt,
	)
	if err != nil {
		return TableSession{}, fmt.Errorf("create table session: %w", err)
	}
	session.TableID = tableID
	if err := tx.QueryRow(ctx, `
		SELECT table_number, COALESCE(display_name, ''), qr_token
		FROM restaurant_tables
		WHERE id = $1`, tableID).Scan(&session.TableNumber, &session.DisplayName, &session.QRToken); err != nil {
		return TableSession{}, fmt.Errorf("load table session details: %w", err)
	}
	return session, nil
}

func generateTableSessionToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
