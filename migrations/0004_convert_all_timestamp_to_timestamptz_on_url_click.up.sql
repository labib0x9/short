-- urls table
ALTER TABLE urls ALTER COLUMN last_clicked_at TYPE timestamptz USING last_clicked_at AT TIME ZONE 'UTC';
ALTER TABLE urls ALTER COLUMN created_at TYPE timestamptz USING created_at AT TIME ZONE 'UTC';
ALTER TABLE urls ALTER COLUMN expire_at TYPE timestamptz USING expire_at AT TIME ZONE 'UTC';

-- clicks table
ALTER TABLE clicks ALTER COLUMN clicked_at TYPE timestamptz USING clicked_at AT TIME ZONE 'UTC';