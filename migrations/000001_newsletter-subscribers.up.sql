CREATE TABLE IF NOT EXISTS newsletter_subscribers (
    id    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    token TEXT not null,
    confirmed BOOLEAN not null default false,
    active BOOLEAN not null default true,
    created_at TIMESTAMPTZ not null default now(),
    updated_at TIMESTAMPTZ not null default now()
);
