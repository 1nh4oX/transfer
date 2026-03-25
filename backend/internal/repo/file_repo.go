package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type FileRecord struct {
	ID          string
	OwnerID     string
	FolderID    *string
	Name        string
	StoragePath string
	MimeType    string
	Ext         string
	Size        int64
	CreatedAt   time.Time
}

type CreateFileParams struct {
	ID          string
	OwnerID     string
	FolderID    *string
	Name        string
	StoragePath string
	MimeType    string
	Ext         string
	Size        int64
	CreatedAt   time.Time
}

type FileRepository interface {
	InitSchema(ctx context.Context) error
	CreateFile(ctx context.Context, params CreateFileParams) error
	GetFileByIDAndOwner(ctx context.Context, fileID, ownerID string) (FileRecord, error)
	GetFileByID(ctx context.Context, fileID string) (FileRecord, error)
	ListFilesByOwner(ctx context.Context, ownerID string, folderID *string, limit, offset int) ([]FileRecord, int64, error)
	DeleteFileByIDAndOwner(ctx context.Context, fileID, ownerID string) (FileRecord, error)
}

type PostgresFileRepository struct {
	db *sql.DB
}

func NewPostgresFileRepository(db *sql.DB) *PostgresFileRepository {
	return &PostgresFileRepository{db: db}
}

func (r *PostgresFileRepository) InitSchema(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS files (
  id TEXT PRIMARY KEY,
  owner_id TEXT NOT NULL,
  folder_id TEXT NULL,
  name TEXT NOT NULL,
  storage_path TEXT NOT NULL,
  mime_type TEXT NOT NULL,
  ext TEXT NOT NULL,
  size BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_files_owner_id ON files (owner_id);
`
	if _, err := r.db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("init files schema: %w", err)
	}
	return nil
}

func (r *PostgresFileRepository) CreateFile(ctx context.Context, params CreateFileParams) error {
	const q = `
INSERT INTO files (
  id, owner_id, folder_id, name, storage_path, mime_type, ext, size, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
`
	_, err := r.db.ExecContext(ctx, q,
		params.ID,
		params.OwnerID,
		params.FolderID,
		params.Name,
		params.StoragePath,
		params.MimeType,
		params.Ext,
		params.Size,
		params.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert file: %w", err)
	}
	return nil
}

func (r *PostgresFileRepository) GetFileByIDAndOwner(ctx context.Context, fileID, ownerID string) (FileRecord, error) {
	const q = `
SELECT id, owner_id, folder_id, name, storage_path, mime_type, ext, size, created_at
FROM files
WHERE id = $1 AND owner_id = $2;
`
	return r.scanOne(ctx, q, fileID, ownerID)
}

func (r *PostgresFileRepository) GetFileByID(ctx context.Context, fileID string) (FileRecord, error) {
	const q = `
SELECT id, owner_id, folder_id, name, storage_path, mime_type, ext, size, created_at
FROM files
WHERE id = $1;
`
	return r.scanOne(ctx, q, fileID)
}

func (r *PostgresFileRepository) ListFilesByOwner(ctx context.Context, ownerID string, folderID *string, limit, offset int) ([]FileRecord, int64, error) {
	const countRoot = `SELECT COUNT(*) FROM files WHERE owner_id = $1 AND folder_id IS NULL;`
	const countWithFolder = `SELECT COUNT(*) FROM files WHERE owner_id = $1 AND folder_id = $2;`
	const listRoot = `
SELECT id, owner_id, folder_id, name, storage_path, mime_type, ext, size, created_at
FROM files
WHERE owner_id = $1 AND folder_id IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;`
	const listWithFolder = `
SELECT id, owner_id, folder_id, name, storage_path, mime_type, ext, size, created_at
FROM files
WHERE owner_id = $1 AND folder_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;`

	var total int64
	var rows *sql.Rows
	var err error

	if folderID == nil {
		if err = r.db.QueryRowContext(ctx, countRoot, ownerID).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("count files: %w", err)
		}
		rows, err = r.db.QueryContext(ctx, listRoot, ownerID, limit, offset)
	} else {
		if err = r.db.QueryRowContext(ctx, countWithFolder, ownerID, *folderID).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("count files: %w", err)
		}
		rows, err = r.db.QueryContext(ctx, listWithFolder, ownerID, *folderID, limit, offset)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("list files: %w", err)
	}
	defer rows.Close()

	items := make([]FileRecord, 0)
	for rows.Next() {
		var rec FileRecord
		if err := rows.Scan(
			&rec.ID,
			&rec.OwnerID,
			&rec.FolderID,
			&rec.Name,
			&rec.StoragePath,
			&rec.MimeType,
			&rec.Ext,
			&rec.Size,
			&rec.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan file: %w", err)
		}
		items = append(items, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return items, total, nil
}

func (r *PostgresFileRepository) DeleteFileByIDAndOwner(ctx context.Context, fileID, ownerID string) (FileRecord, error) {
	const q = `
DELETE FROM files
WHERE id = $1 AND owner_id = $2
RETURNING id, owner_id, folder_id, name, storage_path, mime_type, ext, size, created_at;
`
	var rec FileRecord
	if err := r.db.QueryRowContext(ctx, q, fileID, ownerID).Scan(
		&rec.ID,
		&rec.OwnerID,
		&rec.FolderID,
		&rec.Name,
		&rec.StoragePath,
		&rec.MimeType,
		&rec.Ext,
		&rec.Size,
		&rec.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return FileRecord{}, ErrNotFound
		}
		return FileRecord{}, fmt.Errorf("delete file: %w", err)
	}
	return rec, nil
}

func (r *PostgresFileRepository) scanOne(ctx context.Context, q string, args ...any) (FileRecord, error) {
	var rec FileRecord
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(
		&rec.ID,
		&rec.OwnerID,
		&rec.FolderID,
		&rec.Name,
		&rec.StoragePath,
		&rec.MimeType,
		&rec.Ext,
		&rec.Size,
		&rec.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return FileRecord{}, ErrNotFound
		}
		return FileRecord{}, fmt.Errorf("query file: %w", err)
	}
	return rec, nil
}
