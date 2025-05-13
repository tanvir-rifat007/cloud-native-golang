CREATE TABLE newsletter_files (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    newsletter_id BIGINT REFERENCES newsletters(id) ON DELETE CASCADE,
    file_url TEXT NOT NULL,
    filename TEXT NOT NULL,
    uploaded_at TIMESTAMPTZ DEFAULT now()
);
