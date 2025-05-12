ALTER TABLE newsletters
ADD COLUMN created_by BIGINT REFERENCES newsletter_subscribers(id) ON DELETE CASCADE;
