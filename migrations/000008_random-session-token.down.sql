DROP TABLE IF EXISTS token;
CREATE TABLE token
(
    id      UUID PRIMARY KEY,
    user_id UUID,
    type    token_type,
    created TIMESTAMPTZ,
    expires TIMESTAMPTZ
);
