ALTER TABLE IF EXISTS expenditure
    ADD COLUMN IF NOT EXISTS created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS source VARCHAR(50);