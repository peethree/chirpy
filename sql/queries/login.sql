-- name: Login :one

SELECT * FROM users WHERE email = $1;