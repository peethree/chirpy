-- name: UpdateUser :exec

UPDATE users
SET email = $1,
    hashed_password = $2,
    updated_at = NOW()
WHERE id = $3;