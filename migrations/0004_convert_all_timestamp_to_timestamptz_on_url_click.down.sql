-- url_click table
ALTER TABLE clicks ALTER COLUMN clicked_at TYPE timestamp USING clicked_at AT TIME ZONE 'UTC';

-- urls table
ALTER TABLE urls ALTER COLUMN expire_at TYPE timestamp USING expire_at AT TIME ZONE 'UTC';
ALTER TABLE urls ALTER COLUMN created_at TYPE timestamp USING created_at AT TIME ZONE 'UTC';
ALTER TABLE urls ALTER COLUMN last_clicked_at TYPE timestamp USING last_clicked_at AT TIME ZONE 'UTC';