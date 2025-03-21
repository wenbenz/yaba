DO $$ BEGIN
    CREATE TYPE token_type AS ENUM ('SESSION');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS token
(
    id      UUID PRIMARY KEY,
    user_id UUID,
    type    token_type,
    created TIMESTAMPTZ,
    expires TIMESTAMPTZ
);
