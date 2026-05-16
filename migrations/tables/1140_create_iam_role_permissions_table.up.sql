CREATE TABLE iam.role_permissions (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    role_id BIGINT NOT NULL REFERENCES iam.roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES iam.permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (role_id, permission_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_iam_role_permissions_uuid ON iam.role_permissions (uuid);
CREATE INDEX IF NOT EXISTS idx_iam_role_permissions_role_id ON iam.role_permissions (role_id);
CREATE INDEX IF NOT EXISTS idx_iam_role_permissions_permission_id ON iam.role_permissions (permission_id);
CREATE INDEX IF NOT EXISTS idx_iam_role_permissions_id ON iam.role_permissions (id);
