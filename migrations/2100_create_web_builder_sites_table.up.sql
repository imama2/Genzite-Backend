CREATE TABLE web_builder.web_builder_sites (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    slug TEXT NOT NULL UNIQUE,
    title TEXT,
    config JSONB NOT NULL,
    status TEXT NOT NULL,
    output_path TEXT NOT NULL,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_web_builder_sites_uuid ON web_builder.web_builder_sites (uuid);
CREATE INDEX IF NOT EXISTS idx_web_builder_sites_user_id ON web_builder.web_builder_sites (user_id);
CREATE INDEX IF NOT EXISTS idx_web_builder_sites_status ON web_builder.web_builder_sites (status);
CREATE INDEX IF NOT EXISTS idx_web_builder_sites_deleted_at ON web_builder.web_builder_sites (deleted_at);
CREATE INDEX IF NOT EXISTS idx_web_builder_sites_id ON web_builder.web_builder_sites (id);
