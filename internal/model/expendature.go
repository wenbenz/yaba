package model

import (
	"time"

	"github.com/google/uuid"
)

type Expenditure struct {
	ID             int       `db:"id"`
	Owner          uuid.UUID `db:"owner"`
	Name           string    `db:"name"`
	Amount         float64   `db:"amount"`
	Date           time.Time `db:"date"`
	Method         uuid.UUID `db:"method"`
	BudgetCategory string    `db:"budget_category"`
	RewardCategory string    `db:"reward_category"`
	Comment        string    `db:"comment"`
	CreatedTime    time.Time `db:"created"`
	Source         string    `db:"source"`
	ExpenseID      uuid.UUID `db:"expense_id"`
}

type ExpenditureSummary struct {
	Category  string    `db:"category"`
	Amount    float64   `db:"amount"`
	StartDate time.Time `db:"date"`
}

type Aggregation string

const (
	AggregationSum     Aggregation = "SUM"
	AggregationAverage Aggregation = "AVG"
)

type Timespan string

const (
	TimespanDay   Timespan = "DAY"
	TimespanWeek  Timespan = "WEEK"
	TimespanMonth Timespan = "MONTH"
	TimespanYear  Timespan = "YEAR"
)

type GroupBy string

const (
	GroupByNone           GroupBy = "NONE"
	GroupByBudgetCategory GroupBy = "BUDGET_CATEGORY"
	GroupByRewardCategory GroupBy = "REWARD_CATEGORY"
)
