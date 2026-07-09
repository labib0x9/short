-- ─────────────────────────────────────────
-- 1. SEED urls (10k rows)
-- ─────────────────────────────────────────
INSERT INTO urls (id, url, short, total, last_clicked_at, created_at, expire_at)
SELECT
    uuid_generate_v4(),

    'https://' ||
    (ARRAY['google','github','youtube','reddit','twitter','amazon','netflix','medium'])[floor(random()*8+1)::int] ||
    '.com/' ||
    md5(random()::text) ||
    '/' ||
    md5(random()::text),

    substring(md5(i::text || random()::text), 1, 8),

    floor(random() * 10000)::bigint,

    NOW() - (random() * INTERVAL '30 days'),

    NOW() - (random() * INTERVAL '180 days'),

    CASE
        WHEN random() < 0.3
        THEN NOW() + (random() * INTERVAL '90 days')
        ELSE NULL
    END

FROM generate_series(1, 10000) AS s(i);

-- ─────────────────────────────────────────
-- 2. SEED clicks (10M rows) — CORRECTED
-- ─────────────────────────────────────────
CREATE TEMP TABLE _url_ids AS
SELECT id, row_number() OVER () AS rn
FROM urls;
CREATE INDEX ON _url_ids(rn);

DO $$
DECLARE
    url_count INT;
BEGIN
    SELECT COUNT(*) INTO url_count FROM _url_ids;
    RAISE NOTICE 'url_count: %', url_count;
END $$;

INSERT INTO clicks (url_id, referer, country, device, os, browser, clicked_at)
SELECT
    u.id,
    (ARRAY[
        'https://google.com',
        'https://twitter.com',
        'https://github.com',
        'https://reddit.com',
        'https://youtube.com',
        'https://facebook.com',
        ''
    ])[floor(random()*7+1)::int],
    (ARRAY[
        'BD','US','UK','IN','DE','FR','JP','BR','CA','AU','NG','PK'
    ])[floor(random()*12+1)::int],
    (ARRAY['Mobile','Desktop','Tablet'])[floor(random()*3+1)::int],
    (ARRAY['Windows','macOS','Linux','Android','iOS'])[floor(random()*5+1)::int],
    (ARRAY['Chrome','Firefox','Safari','Edge','Opera'])[floor(random()*5+1)::int],
    NOW() - random() * INTERVAL '180 days'
FROM (
    SELECT floor(random() * 10000 + 1)::int AS rn
    FROM generate_series(1, 10000000)
) g
JOIN _url_ids u
ON u.rn = g.rn;

DROP TABLE _url_ids;