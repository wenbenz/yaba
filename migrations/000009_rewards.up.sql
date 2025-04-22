ALTER TABLE IF EXISTS rewards_card
    DROP COLUMN IF EXISTS reward_rate,
    DROP COLUMN IF EXISTS reward_cash_value;

CREATE TABLE IF NOT EXISTS card_rewards(
    card_id UUID NOT NULL,
    category VARCHAR(50) NOT NULL,
    reward_rate NUMERIC(6, 4) NOT NULL,

    PRIMARY KEY (card_id, category),
    FOREIGN KEY (card_id) REFERENCES rewards_card(id)
);
