package budget

import (
	"strings"

	"github.com/google/uuid"
)

type Strategy uint8

const (
	zeroBased Strategy = iota
)

type Budget struct {
	ID       uuid.UUID `db:"id"`
	Name     string    `db:"name"`
	Strategy Strategy  `db:"strategy"`
	Incomes  map[string]*Income
	Expenses map[string]*Expense
}

type Income struct {
	Owner  uuid.UUID `db:"owner"`
	Source string    `db:"source"`
	Amount uint      `db:"amount"`
}

type Expense struct {
	BudgetID uuid.UUID `db:"budget_id"`
	Category string    `db:"category"`
	Amount   uint      `db:"amount"`
	Fixed    bool      `db:"is_fixed"`
	Slack    bool      `db:"is_slack"`
}

func NewZeroBasedBudget(name string) *Budget {
	return &Budget{
		ID:       uuid.New(),
		Name:     name,
		Strategy: zeroBased,
		Incomes:  make(map[string]*Income),
		Expenses: make(map[string]*Expense),
	}
}

func (b *Budget) SetBudgetIncome(source string, amount uint) {
	s := strings.ToLower(source)
	b.Incomes[s] = &Income{
		Owner:  b.ID,
		Source: source,
		Amount: amount,
	}
}

func (b *Budget) RemoveBudgetIncome(source string) {
	s := strings.ToLower(source)
	delete(b.Incomes, s)
}

func (b *Budget) SetFixedExpense(category string, amount uint) {
	b.SetExpense(category, amount, true, false)
}

func (b *Budget) SetPercentageExpense(category string, amount uint) {
	b.SetExpense(category, amount, false, false)
}

func (b *Budget) SetSlackExpense(category string) {
	b.SetExpense(category, 0, false, true)
}

func (b *Budget) SetExpense(category string, amount uint, fixed, slack bool) {
	c := strings.ToLower(category)
	b.Expenses[c] = &Expense{
		BudgetID: b.ID,
		Category: c,
		Amount:   amount,
		Fixed:    fixed,
		Slack:    slack,
	}
}

func (b *Budget) RemoveExpense(category string) {
	c := strings.ToLower(category)
	delete(b.Expenses, c)
}
