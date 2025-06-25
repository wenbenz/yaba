ALTER TABLE expenditure
    ALTER COLUMN budget_category TYPE VARCHAR(50) USING SUBSTRING(budget_category FOR 50);