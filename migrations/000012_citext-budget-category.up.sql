CREATE EXTENSION citext;
ALTER TABLE expenditure
    ALTER COLUMN budget_category TYPE CITEXT USING budget_category::CITEXT;