CREATE TABLE iam.roles (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_iam_roles_uuid ON iam.roles (uuid);
CREATE INDEX IF NOT EXISTS idx_iam_roles_deleted_at ON iam.roles (deleted_at);
CREATE INDEX IF NOT EXISTS idx_iam_roles_id ON iam.roles (id);
