-- +goose Up
CREATE TABLE chirps (
    id uuid PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    body text NOT NULL,
    user_id uuid NOT NULL,
    CONSTRAINT fk_users FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- +goose Down
-- pgls-ignore lint/safety/banDropTable: down migration
DROP TABLE chirps;
