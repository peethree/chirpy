-- name: GetUserIdFromChirp :one
SELECT * FROM chirps WHERE id = $1;