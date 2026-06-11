CREATE TABLE url_stats (
    url_id          UUID           NOT NULL UNIQUE REFERENCES urls(id) ON DELETE CASCADE,
    total           BIGINT    NOT NULL DEFAULT 0,
    last_clicked_at TIMESTAMP  NOT NULL
);