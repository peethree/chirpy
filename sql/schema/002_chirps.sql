-- +goose Up
CREATE TABLE chirps (
-- A new random id: A UUID
    id UUID PRIMARY KEY,
-- created_at: A non-null timestamp
    created_at TIMESTAMP NOT NULL,
-- updated_at: A non null timestamp
    updated_at TIMESTAMP NOT NULL,
-- body: A non-null string
    body TEXT NOT NULL,
-- user_id:
    user_id UUID NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
); 

-- +goose Down
DROP TABLE chirps;

-- postgres://postgres:postgres@localhost:5432/chirpy