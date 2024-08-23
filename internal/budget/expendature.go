package budget

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
	Method         string    `db:"method"`
	BudgetCategory string    `db:"budget_category"`
	RewardCategory string    `db:"reward_category"`
	Comment        string    `db:"comment"`
	CreatedTime    time.Time `db:"created"`
	Source         string    `db:"source"`
}

type ExpenditureSummary struct {
	Category  string    `db:"category"`
	Amount    float64   `db:"amount"`
	StartDate time.Time `db:"date"`
}
