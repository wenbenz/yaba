ALTER TABLE expenditure
    ALTER COLUMN reward_category TYPE VARCHAR(30)
        USING LOWER(REPLACE(reward_category::text, '_', ' ')),
    ALTER COLUMN reward_category SET DEFAULT '';

DROP TYPE IF EXISTS reward_category;
