CREATE SCHEMA palo_verde;

CREATE TABLE IF NOT EXISTS palo_verde.user (
    id uuid PRIMARY KEY,
    username character varying(255) NOT NULL,
    created timestamp without time zone NOT NULL,
    last_seen timestamp without time zone NOT NULL
);