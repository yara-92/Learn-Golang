package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/yorkyu/approval-engine/internal/model"
)

func CreateUser(ctx context.Context, q Querier, u *model.User) (int64, error) {
	res, err := q.ExecContext(ctx,
		`INSERT INTO users (username, password_hash, display_name, role) VALUES (?, ?, ?, ?)`,
		u.Username, u.PasswordHash, u.DisplayName, u.Role,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetUserByUsername(ctx context.Context, q Querier, username string) (*model.User, error) {
	row := q.QueryRowContext(ctx,
		`SELECT id, username, password_hash, display_name, role, created_at FROM users WHERE username = ?`,
		username,
	)
	return scanUser(row)
}

func GetUserByID(ctx context.Context, q Querier, id int64) (*model.User, error) {
	row := q.QueryRowContext(ctx,
		`SELECT id, username, password_hash, display_name, role, created_at FROM users WHERE id = ?`,
		id,
	)
	return scanUser(row)
}

func ListUsersByRole(ctx context.Context, q Querier, role string) ([]model.User, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT id, username, password_hash, display_name, role, created_at FROM users WHERE role = ? ORDER BY id`,
		role,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func ListAllUsers(ctx context.Context, q Querier) ([]model.User, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT id, username, password_hash, display_name, role, created_at FROM users ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func scanUser(row *sql.Row) (*model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &u.Role, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
