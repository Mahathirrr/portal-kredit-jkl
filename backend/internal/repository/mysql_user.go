package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Mahathirrr/kredit-jkl/internal/domain"
)

func (s Store) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	const q = `SELECT id, username, password_hash, role, nama, created_at FROM users WHERE username = ?`
	row := s.DB.QueryRowContext(ctx, q, username)
	return scanUser(row)
}

func (s Store) FindUserByID(ctx context.Context, id int64) (*domain.User, error) {
	const q = `SELECT id, username, password_hash, role, nama, created_at FROM users WHERE id = ?`
	row := s.DB.QueryRowContext(ctx, q, id)
	return scanUser(row)
}

func (s Store) CreateDemoUsersIfEmpty(ctx context.Context, users []domain.User) error {
	var n int
	if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	const q = `INSERT INTO users (username, password_hash, role, nama) VALUES (?, ?, ?, ?)`
	for _, u := range users {
		if _, err := s.DB.ExecContext(ctx, q, u.Username, u.PasswordHash, u.Role, u.Nama); err != nil {
			return err
		}
	}
	return nil
}

func scanUser(row *sql.Row) (*domain.User, error) {
	var u domain.User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.Nama, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}