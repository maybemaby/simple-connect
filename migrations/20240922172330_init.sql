-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY NOT NULL GENERATED ALWAYS AS IDENTITY,
    password_hash TEXT,
    email TEXT UNIQUE,
    email_verified TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY NOT NULL,
    expiry TIMESTAMPTZ NOT NULL,
    data BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

CREATE TABLE IF NOT EXISTS accounts (
    id INT PRIMARY KEY NOT NULL GENERATED ALWAYS AS IDENTITY,
    provider_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    user_id INT REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    access_token TEXT,
    refresh_token TEXT,
    access_token_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_account_user_provider ON accounts (provider, user_id);
CREATE UNIQUE INDEX accounts_provider_provider_id_idx ON accounts (provider, provider_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE sessions;
DROP TABLE accounts;
DROP TABLE users;
-- +goose StatementEnd