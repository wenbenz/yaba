/* Sources:
 *  - https://forums.redflagdeals.com/ultimate-tangerine-mc-category-reference-guide-2278771/
 *  - https://www.tangerine.ca/en/personal/spend/credit-cards/money-back-credit-card
 */
DO $$ BEGIN
    CREATE TYPE reward_category AS ENUM (
        '',
        'DRUG_STORE',
        'ENTERTAINMENT',
        'FURNITURE',
        'GAS',
        'GROCERY',
        'HOME_IMPROVEMENT',
        'HOTEL',
        'PUBLIC_TRANSPORTATION',
        'RECURRING_BILL',
        'RESTAURANT'
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS expenditure (
    id SERIAL PRIMARY KEY,
    owner UUID NOT NULL,
    amount NUMERIC(20,4) NOT NULL,
    date DATE NOT NULL,
    name VARCHAR(50),
    method VARCHAR(50),
    budget_category VARCHAR(50),
    reward_category reward_category NOT NULL DEFAULT '',
    comment TEXT
);

CREATE INDEX IF NOT EXISTS idx_owner_date_id ON expenditure USING BTREE(owner, date, id);
