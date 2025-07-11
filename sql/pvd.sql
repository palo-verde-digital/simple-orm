CREATE SCHEMA IF NOT EXISTS palo_verde;

CREATE TABLE IF NOT EXISTS palo_verde.user (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL,
    logins INTEGER NOT NULL,
    created TIMESTAMP NOT NULL,
    last_seen TIMESTAMP NOT NULL
);