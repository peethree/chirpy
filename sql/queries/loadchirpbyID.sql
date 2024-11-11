-- name: LoadChirpByID :one
SELECT * FROM chirps WHERE id = $1; 