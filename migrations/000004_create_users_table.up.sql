CREATE TABLE IF NOT EXISTS "user"
(
    id           UUID PRIMARY KEY NOT NULL,
    username     VARCHAR(50)      NOT NULL UNIQUE,
    password_hash BYTEA
);
