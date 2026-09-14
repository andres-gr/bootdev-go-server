-- +goose Up
-- +goose StatementBegin
SET local lock_timeout = '3s';

SET local statement_timeout = '5s';

ALTER TABLE users
    ADD COLUMN is_chirpy_red boolean NOT NULL DEFAULT FALSE;

-- +goose StatementEnd

-- +goose Down
-- pgls-ignore-start lint/safety/banDropColumn
-- pgls-ignore-start lint/safety/multipleAlterTable
-- pgls-ignore-start lint/safety/runningStatementWhileHoldingAccessExclusive
ALTER TABLE users
    DROP COLUMN is_chirpy_red;

-- pgls-ignore-end lint/safety/banDropColumn
-- pgls-ignore-end lint/safety/multipleAlterTable
-- pgls-ignore-end lint/safety/runningStatementWhileHoldingAccessExclusive
