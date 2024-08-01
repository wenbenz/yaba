CREATE TABLE IF NOT EXISTS budget (
    id UUID PRIMARY KEY,
    name VARCHAR(50),
    strategy SMALLINT
);

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