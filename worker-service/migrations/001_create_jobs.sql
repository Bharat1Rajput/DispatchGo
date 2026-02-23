
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY,

    payload TEXT NOT NULL,
    target_url TEXT NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'pending',

    retry_count INTEGER NOT NULL DEFAULT 0,

    error TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT jobs_status_check
        CHECK (status IN ('pending', 'processing', 'completed', 'failed'))
);

-- Index for polling (monolith phase)
CREATE INDEX IF NOT EXISTS idx_jobs_status
    ON jobs(status);

-- Optional but useful
CREATE INDEX IF NOT EXISTS idx_jobs_created_at
    ON jobs(created_at);