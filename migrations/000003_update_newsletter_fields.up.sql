ALTER TABLE newsletter_subscribers
ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT false;
