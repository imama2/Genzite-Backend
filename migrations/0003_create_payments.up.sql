CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    site_id BIGINT NOT NULL REFERENCES web_builder_sites(id) ON DELETE CASCADE,
    order_id TEXT NOT NULL UNIQUE,
    amount BIGINT NOT NULL,
    status TEXT NOT NULL,
    provider TEXT NOT NULL,
    redirect_url TEXT,
    snap_token TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_payments_user_id ON payments (user_id);
CREATE INDEX idx_payments_site_id ON payments (site_id);
CREATE INDEX idx_payments_status ON payments (status);
CREATE INDEX idx_payments_deleted_at ON payments (deleted_at);
