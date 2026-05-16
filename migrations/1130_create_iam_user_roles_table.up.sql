CREATE TABLE iam.user_roles (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES iam.roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (user_id, role_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_iam_user_roles_uuid ON iam.user_roles (uuid);
CREATE INDEX IF NOT EXISTS idx_iam_user_roles_user_id ON iam.user_roles (user_id);
CREATE INDEX IF NOT EXISTS idx_iam_user_roles_role_id ON iam.user_roles (role_id);
CREATE INDEX IF NOT EXISTS idx_iam_user_roles_id ON iam.user_roles (id);
