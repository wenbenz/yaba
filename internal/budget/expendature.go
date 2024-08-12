package budget

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Expenditure struct {
	Owner          uuid.UUID `db:"owner"`
	Name           string    `db:"name"`
	Amount         float64   `db:"amount"`
	Date           time.Time `db:"date"`
	Method         string    `db:"method"`
	BudgetCategory string    `db:"budget_category"`
	RewardCategory sql.NullString    `db:"reward_category"`
	Comment        string    `db:"comment"`
	ID             int       `db:"id"`
}
