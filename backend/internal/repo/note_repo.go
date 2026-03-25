package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type NoteRecord struct {
	ID        string
	OwnerID   string
	Content   string
	CreatedAt time.Time
}

type CreateNoteParams struct {
	ID        string
	OwnerID   string
	Content   string
	CreatedAt time.Time
}

type NoteRepository interface {
	InitSchema(ctx context.Context) error
	CreateNote(ctx context.Context, params CreateNoteParams) error
	ListNotesByOwner(ctx context.Context, ownerID string, limit, offset int) ([]NoteRecord, int64, error)
}

type PostgresNoteRepository struct {
	db *sql.DB
}

func NewPostgresNoteRepository(db *sql.DB) *PostgresNoteRepository {
	return &PostgresNoteRepository{db: db}
}

func (r *PostgresNoteRepository) InitSchema(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS notes (
  id TEXT PRIMARY KEY,
  owner_id TEXT NOT NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_notes_owner_id_created_at ON notes (owner_id, created_at DESC);
`
	if _, err := r.db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("init notes schema: %w", err)
	}
	return nil
}

func (r *PostgresNoteRepository) CreateNote(ctx context.Context, params CreateNoteParams) error {
	const q = `
INSERT INTO notes (id, owner_id, content, created_at)
VALUES ($1, $2, $3, $4);`
	if _, err := r.db.ExecContext(ctx, q, params.ID, params.OwnerID, params.Content, params.CreatedAt); err != nil {
		return fmt.Errorf("insert note: %w", err)
	}
	return nil
}

func (r *PostgresNoteRepository) ListNotesByOwner(ctx context.Context, ownerID string, limit, offset int) ([]NoteRecord, int64, error) {
	const countQ = `SELECT COUNT(*) FROM notes WHERE owner_id = $1;`
	const listQ = `
SELECT id, owner_id, content, created_at
FROM notes
WHERE owner_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;`

	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, ownerID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count notes: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, listQ, ownerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	items := make([]NoteRecord, 0)
	for rows.Next() {
		var rec NoteRecord
		if err := rows.Scan(&rec.ID, &rec.OwnerID, &rec.Content, &rec.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan note: %w", err)
		}
		items = append(items, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return items, total, nil
}
