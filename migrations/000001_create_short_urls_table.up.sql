CREATE TABLE short_urls (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(20) NOT NULL,
    original_url TEXT NOT NULL,
    url_hash CHAR(64),
    user_id BIGINT,
    expires_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_short_urls_code ON short_urls (code);
CREATE INDEX idx_short_urls_url_hash ON short_urls (url_hash);
CREATE INDEX idx_short_urls_expires_at ON short_urls (expires_at);
