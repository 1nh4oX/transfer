package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrNotFound = errors.New("not found")

type ShareRecord struct {
	Token     string
	OwnerID   string
	ItemType  string
	ItemID    string
	ItemName  string
	ItemSize  *int64
	CreatedAt time.Time
	ExpiresAt *time.Time
}

type CreateShareParams struct {
	Token     string
	OwnerID   string
	ItemType  string
	ItemID    string
	ItemName  string
	ItemSize  *int64
	CreatedAt time.Time
	ExpiresAt *time.Time
}

type ShareRepository interface {
	CreateShare(ctx context.Context, params CreateShareParams) error
	GetShareByToken(ctx context.Context, token string) (ShareRecord, error)
}

type PostgresShareRepository struct {
	db *sql.DB
}

func NewPostgresDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return db, nil
}

func NewPostgresShareRepository(db *sql.DB) *PostgresShareRepository {
	return &PostgresShareRepository{db: db}
}

func (r *PostgresShareRepository) InitSchema(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS shares (
  token TEXT PRIMARY KEY,
  owner_id TEXT NOT NULL,
  item_type TEXT NOT NULL CHECK (item_type IN ('file', 'folder')),
  item_id TEXT NOT NULL,
  item_name TEXT NOT NULL,
  item_size BIGINT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  expires_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS idx_shares_expires_at ON shares (expires_at);
`
	if _, err := r.db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("init shares schema: %w", err)
	}
	return nil
}

func (r *PostgresShareRepository) CreateShare(ctx context.Context, params CreateShareParams) error {
	const q = `
INSERT INTO shares (
  token, owner_id, item_type, item_id, item_name, item_size, created_at, expires_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
`
	_, err := r.db.ExecContext(
		ctx,
		q,
		params.Token,
		params.OwnerID,
		params.ItemType,
		params.ItemID,
		params.ItemName,
		params.ItemSize,
		params.CreatedAt,
		params.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert share: %w", err)
	}

	return nil
}

func (r *PostgresShareRepository) GetShareByToken(ctx context.Context, token string) (ShareRecord, error) {
	const q = `
SELECT token, owner_id, item_type, item_id, item_name, item_size, created_at, expires_at
FROM shares
WHERE token = $1;
`

	var rec ShareRecord
	if err := r.db.QueryRowContext(ctx, q, token).Scan(
		&rec.Token,
		&rec.OwnerID,
		&rec.ItemType,
		&rec.ItemID,
		&rec.ItemName,
		&rec.ItemSize,
		&rec.CreatedAt,
		&rec.ExpiresAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ShareRecord{}, ErrNotFound
		}
		return ShareRecord{}, fmt.Errorf("query share: %w", err)
	}

	return rec, nil
}
