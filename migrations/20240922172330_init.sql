-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    password_hash TEXT,
    email TEXT UNIQUE,
    email_verified TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY NOT NULL,
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    expiry TIMESTAMPTZ NOT NULL,
    data BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

CREATE TABLE IF NOT EXISTS accounts (
    provider_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    access_token TEXT,
    refresh_token TEXT,
    access_token_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (provider, provider_id)
);

CREATE UNIQUE INDEX idx_account_user_provider ON accounts (provider, user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE sessions;
DROP TABLE accounts;
DROP TABLE users;
-- +goose StatementEnd