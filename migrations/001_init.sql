CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY,
    title VARCHAR(120) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_items_created_at_desc
    ON items (created_at DESC);
