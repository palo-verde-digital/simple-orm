CREATE SCHEMA pvd_test;

CREATE TABLE IF NOT EXISTS pvd_test.user (
    id uuid PRIMARY KEY,
    username character varying(255) NOT NULL,
    created timestamp without time zone NOT NULL,
    last_seen timestamp without time zone NOT NULL
);