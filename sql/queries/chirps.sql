-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
    VALUES (uuidv7(), now(), now(), $1, $2)
RETURNING
    *;
