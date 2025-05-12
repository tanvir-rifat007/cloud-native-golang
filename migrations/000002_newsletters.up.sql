CREATE TABLE IF NOT EXISTS newsletters (
    id    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title TEXT not null,
    body  TEXT not null,
    tags TEXT[] DEFAULT '{}',
 
    created_at TIMESTAMPTZ not null default now(),
    updated_at TIMESTAMPTZ not null default now()
);
