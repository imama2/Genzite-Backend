CREATE TABLE iam.permissions (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_iam_permissions_uuid ON iam.permissions (uuid);
CREATE INDEX IF NOT EXISTS idx_iam_permissions_deleted_at ON iam.permissions (deleted_at);
CREATE INDEX IF NOT EXISTS idx_iam_permissions_id ON iam.permissions (id);
