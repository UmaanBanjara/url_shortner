CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    short_code TEXT NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE clicks (
    id BIGSERIAL PRIMARY KEY,
    url_id BIGINT NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
    user_agent TEXT,
    ip_address TEXT,
    clicked_at TIMESTAMP DEFAULT NOW()
);