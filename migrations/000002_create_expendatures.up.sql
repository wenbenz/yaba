CREATE TABLE IF NOT EXISTS expenditure (
    id SERIAL PRIMARY KEY,
    owner UUID NOT NULL,
    amount NUMERIC(20,4) NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    name VARCHAR(50),
    method VARCHAR(50),
    budget_category VARCHAR(20),
    reward_category VARCHAR(20),
    comment TEXT
);

CREATE INDEX IF NOT EXISTS idx_owner_date ON expenditure USING BTREE(owner, date);
