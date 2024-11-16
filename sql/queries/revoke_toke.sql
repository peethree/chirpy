-- name: RevokeToken :exec

UPDATE refresh_tokens
SET updated_at = NOW(),
    revoked_at = NOW(),
    token = 'revoked'
WHERE token = $1;