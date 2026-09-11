-- +goose Up
-- +goose StatementBegin
SET local lock_timeout = '3s';

SET local statement_timeout = '5s';

ALTER TABLE users
    ADD COLUMN hashed_password text NOT NULL DEFAULT 'unset';

-- +goose StatementEnd

-- +goose Down
-- pgls-ignore-start lint/safety/banDropColumn
-- pgls-ignore-start lint/safety/multipleAlterTable
-- pgls-ignore-start lint/safety/runningStatementWhileHoldingAccessExclusive
ALTER TABLE users
    DROP COLUMN hashed_password;

-- pgls-ignore-end lint/safety/banDropColumn
-- pgls-ignore-end lint/safety/multipleAlterTable
-- pgls-ignore-end lint/safety/runningStatementWhileHoldingAccessExclusive
