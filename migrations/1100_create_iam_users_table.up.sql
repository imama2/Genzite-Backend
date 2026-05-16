CREATE TABLE iam.users (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    name TEXT,
    password_hash TEXT,
    google_id TEXT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_iam_users_uuid ON iam.users (uuid);
CREATE INDEX IF NOT EXISTS idx_iam_users_deleted_at ON iam.users (deleted_at);
CREATE INDEX IF NOT EXISTS idx_iam_users_id ON iam.users (id);
