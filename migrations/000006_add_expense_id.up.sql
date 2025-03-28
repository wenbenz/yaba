CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

ALTER TABLE IF EXISTS expense
    ADD COLUMN IF NOT EXISTS id UUID UNIQUE DEFAULT UUID_GENERATE_V4();

ALTER TABLE IF EXISTS expenditure
    ADD COLUMN IF NOT EXISTS expense_id UUID;

UPDATE expenditure e
    SET expense_id = exp.id
    FROM expense exp
             JOIN budget b ON exp.budget_id = b.id
    WHERE b.owner = e.owner
      AND LOWER(exp.category) = LOWER(e.budget_category);
