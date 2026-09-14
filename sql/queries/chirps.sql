-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
    VALUES (uuidv7 (), now(), now(), $1, $2)
RETURNING
    *;

-- name: GetChirps :many
SELECT
    *
FROM
    chirps
WHERE
    CASE WHEN @author_id::uuid != '00000000-0000-0000-0000-000000000000' THEN
        user_id = @author_id::uuid
    ELSE
        TRUE
    END
ORDER BY
    CASE WHEN @sort::text = 'asc' THEN
        created_at
    END ASC,
    CASE WHEN @sort::text = 'desc' THEN
        created_at
    END DESC;

-- name: GetChirp :one
SELECT
    *
FROM
    chirps
WHERE
    id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE id = $1;
