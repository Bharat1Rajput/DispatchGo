CREATE TABLE IF NOT EXISTS webhook_jobs (
    id          TEXT        PRIMARY KEY,
    payload     TEXT        NOT NULL,
    client_url  TEXT        NOT NULL,
    status      TEXT        NOT NULL DEFAULT 'pending',
    error       TEXT        NOT NULL DEFAULT '',
    retry_count INTEGER     NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

