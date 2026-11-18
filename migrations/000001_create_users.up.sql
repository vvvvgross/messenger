BEGIN;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT users_username_unique UNIQUE (username),

    CONSTRAINT users_username_format CHECK (
        username ~ '^[a-z0-9_]{3,32}$'
    ),

    CONSTRAINT users_password_hash_not_empty CHECK (
        length(btrim(password_hash)) > 0
    )
);

COMMIT;