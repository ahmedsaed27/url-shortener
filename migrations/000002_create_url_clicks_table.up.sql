CREATE TABLE url_clicks (
    id BIGSERIAL PRIMARY KEY,
    short_url_id BIGINT,
    code VARCHAR(20) NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    referer TEXT,
    clicked_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_url_clicks_code ON url_clicks (code);
CREATE INDEX idx_url_clicks_clicked_at ON url_clicks (clicked_at);
