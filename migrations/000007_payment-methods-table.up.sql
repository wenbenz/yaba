CREATE TABLE IF NOT EXISTS payment_method
(
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner          UUID NOT NULL,
    display_name   text,
    card_type      UUID,
    acquired_date  TIMESTAMP,
    cancel_by_date TIMESTAMP
);

CREATE TABLE IF NOT EXISTS rewards_card
(
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name              TEXT           NOT NULL,
    region            TEXT           NOT NULL,
    issuer            TEXT           NOT NULL,
    version           SERIAL         NOT NULL,
    reward_type       TEXT           NOT NULL,
    reward_rate       DECIMAL(17, 5) NOT NULL,
    reward_cash_value DECIMAL(17, 5) NOT NULL
);
