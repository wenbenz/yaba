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

ALTER TABLE expenditure
    ALTER COLUMN reward_category DROP DEFAULT,
    ALTER COLUMN reward_category TYPE reward_category
        USING UPPER(REPLACE(reward_category::text, ' ', '_'))::reward_category,
    ALTER COLUMN reward_category SET DEFAULT '';
