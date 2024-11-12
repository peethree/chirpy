// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: loadchirpbyID.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const loadChirpByID = `-- name: LoadChirpByID :one
SELECT id, created_at, updated_at, body, user_id FROM chirps WHERE id = $1
`

func (q *Queries) LoadChirpByID(ctx context.Context, id uuid.UUID) (Chirp, error) {
	row := q.db.QueryRowContext(ctx, loadChirpByID, id)
	var i Chirp
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Body,
		&i.UserID,
	)
	return i, err
}