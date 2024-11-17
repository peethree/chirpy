-- name: FindEmail :one
SELECT * FROM users 
WHERE email = $1;