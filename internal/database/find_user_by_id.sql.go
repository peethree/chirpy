// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: find_user_by_id.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const findUserById = `-- name: FindUserById :one
SELECT id, created_at, updated_at, email, hashed_password FROM users WHERE id = $1
`

func (q *Queries) FindUserById(ctx context.Context, id uuid.UUID) (User, error) {
	row := q.db.QueryRowContext(ctx, findUserById, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
		&i.HashedPassword,
	)
	return i, err
}
