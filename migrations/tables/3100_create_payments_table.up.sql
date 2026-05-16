CREATE TABLE payments.payments (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    site_id BIGINT NOT NULL REFERENCES web_builder.web_builder_sites(id) ON DELETE CASCADE,
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

CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_uuid ON payments.payments (uuid);
CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments.payments (user_id);
CREATE INDEX IF NOT EXISTS idx_payments_site_id ON payments.payments (site_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments.payments (status);
CREATE INDEX IF NOT EXISTS idx_payments_deleted_at ON payments.payments (deleted_at);
CREATE INDEX IF NOT EXISTS idx_payments_id ON payments.payments (id);
