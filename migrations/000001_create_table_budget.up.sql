CREATE TABLE IF NOT EXISTS budget (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner UUID NOT NULL,
    name VARCHAR(50)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_owner_id ON budget USING BTREE(owner, id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_owner_name ON budget USING BTREE(owner, name);

CREATE TABLE IF NOT EXISTS income (
    owner UUID,
    source VARCHAR(50),
    amount NUMERIC(20,4),
    PRIMARY KEY (owner, source)
);

CREATE TABLE IF NOT EXISTS expense (
    budget_id UUID,
    category VARCHAR(20),
    amount NUMERIC(20,4),
    is_fixed BOOLEAN,
    is_slack BOOLEAN,
    PRIMARY KEY (budget_id, category)
);
