CREATE TABLE web_builder_sites (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
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

CREATE INDEX idx_web_builder_sites_user_id ON web_builder_sites (user_id);
CREATE INDEX idx_web_builder_sites_status ON web_builder_sites (status);
CREATE INDEX idx_web_builder_sites_deleted_at ON web_builder_sites (deleted_at);
