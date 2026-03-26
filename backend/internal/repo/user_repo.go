package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrAlreadyExists = errors.New("already exists")

type UserRecord struct {
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

type CreateUserParams struct {
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

type UserRepository interface {
	InitSchema(ctx context.Context) error
	CreateUser(ctx context.Context, params CreateUserParams) error
	GetUserByUsername(ctx context.Context, username string) (UserRecord, error)
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) InitSchema(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS users (
  username TEXT PRIMARY KEY,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);
`
	if _, err := r.db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("init users schema: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, params CreateUserParams) error {
	const q = `
INSERT INTO users (username, password_hash, created_at)
VALUES ($1, $2, $3)
ON CONFLICT (username) DO NOTHING;
`
	result, err := r.db.ExecContext(ctx, q, params.Username, params.PasswordHash, params.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrAlreadyExists
	}
	return nil
}

func (r *PostgresUserRepository) GetUserByUsername(ctx context.Context, username string) (UserRecord, error) {
	const q = `
SELECT username, password_hash, created_at
FROM users
WHERE username = $1;
`
	var rec UserRecord
	if err := r.db.QueryRowContext(ctx, q, username).Scan(&rec.Username, &rec.PasswordHash, &rec.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserRecord{}, ErrNotFound
		}
		return UserRecord{}, fmt.Errorf("query user: %w", err)
	}
	return rec, nil
}
