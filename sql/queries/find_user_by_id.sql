-- name: FindUserById :one
SELECT * FROM users WHERE id = $1;