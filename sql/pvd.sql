CREATE SCHEMA IF NOT EXISTS palo_verde;

CREATE TABLE IF NOT EXISTS palo_verde.user (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL,
    logins INTEGER NOT NULL,
    created TIMESTAMP NOT NULL,
    last_seen TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS palo_verde.session (
    id UUID PRIMARY KEY,
    created TIMESTAMP NOT NULL,
    user_id UUID REFERENCES palo_verde.user (id)
);