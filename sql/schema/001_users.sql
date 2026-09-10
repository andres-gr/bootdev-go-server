-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    email text NOT NULL,
    UNIQUE (email)
);

-- +goose Down
-- pgls-ignore lint/safety/banDropTable: down migration
DROP TABLE users;
