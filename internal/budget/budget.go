package budget

import (
	"strings"

	"github.com/google/uuid"
)

type Strategy int8

const (
	Unknown Strategy = iota - 1
	ZeroBased
)

// External representation of Budget.
type ExternalBudget struct {
	ID       uuid.UUID  `json:"id,omitempty"`
	Owner    uuid.UUID  `json:"owner"`
	Name     string     `json:"name"`
	Strategy string     `json:"strategy"`
	Incomes  []*Income  `json:"incomes"`
	Expenses []*Expense `json:"expenses"`
}

type Budget struct {
	ID       uuid.UUID `db:"id"`
	Owner    uuid.UUID `db:"owner"`
	Name     string    `db:"name"`
	Strategy Strategy  `db:"strategy"`
	Incomes  map[string]*Income
	Expenses map[string]*Expense
}

type Income struct {
	Owner  uuid.UUID `db:"owner"  json:"-"`
	Source string    `db:"source" json:"source"`
	Amount float64   `db:"amount" json:"amount"`
}

type Expense struct {
	BudgetID uuid.UUID `db:"budget_id" json:"-"`
	Category string    `db:"category"  json:"category"`
	Amount   float64   `db:"amount"    json:"amount"`
	Fixed    bool      `db:"is_fixed"  json:"isFixed"`
	Slack    bool      `db:"is_slack"  json:"isSlack"`
}

func NewZeroBasedBudget(owner uuid.UUID, name string) *Budget {
	return &Budget{
		ID:       uuid.New(),
		Owner:    owner,
		Name:     name,
		Strategy: ZeroBased,
		Incomes:  make(map[string]*Income),
		Expenses: make(map[string]*Expense),
	}
}

func (b *Budget) SetBudgetIncome(source string, amount float64) {
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

func (b *Budget) SetFixedExpense(category string, amount float64) {
	b.SetExpense(category, amount, true, false)
}

func (b *Budget) SetPercentageExpense(category string, amount float64) {
	b.SetExpense(category, amount, false, false)
}

func (b *Budget) SetSlackExpense(category string) {
	b.SetExpense(category, 0, false, true)
}

func (b *Budget) SetExpense(category string, amount float64, fixed, slack bool) {
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

func (b *Budget) ToExternal() *ExternalBudget {
	var strategy string

	switch b.Strategy {
	case ZeroBased:
		strategy = "ZERO_BASED"
	case Unknown:
		strategy = "UNKNOWN"
	}

	i := 0

	incomes := make([]*Income, len(b.Incomes))
	for _, in := range b.Incomes {
		incomes[i] = in
		i++
	}

	i = 0

	expenses := make([]*Expense, len(b.Expenses))
	for _, ex := range b.Expenses {
		expenses[i] = ex
		i++
	}

	return &ExternalBudget{
		ID:       b.ID,
		Owner:    b.Owner,
		Name:     b.Name,
		Strategy: strategy,
		Incomes:  incomes,
		Expenses: expenses,
	}
}

func (b *ExternalBudget) ToInternal() *Budget {
	var strategy Strategy

	switch b.Strategy {
	case "ZERO_BASED":
		strategy = ZeroBased
	default:
		strategy = Unknown
	}

	incomes := make(map[string]*Income, len(b.Incomes))
	for _, in := range b.Incomes {
		incomes[in.Source] = in
	}

	expenses := make(map[string]*Expense, len(b.Expenses))
	for _, ex := range b.Expenses {
		expenses[ex.Category] = ex
	}

	return &Budget{
		ID:       b.ID,
		Owner:    b.Owner,
		Name:     b.Name,
		Strategy: strategy,
		Incomes:  incomes,
		Expenses: expenses,
	}
}
