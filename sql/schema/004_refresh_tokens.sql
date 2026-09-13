-- +goose Up
CREATE TABLE refresh_tokens (
    token text PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    user_id uuid NOT NULL,
    expires_at timestamp NOT NULL,
    revoked_at timestamp,
    CONSTRAINT fk_users FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- +goose Down
-- pgls-ignore lint/safety/banDropTable: down migration
DROP TABLE refresh_tokens;
