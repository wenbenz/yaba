package model

import (
	"github.com/google/uuid"
)

type Strategy int8

type Budget struct {
	ID       uuid.UUID `db:"id"`
	Owner    uuid.UUID `db:"owner"`
	Name     string    `db:"name"`
	Incomes  []*Income
	Expenses []*Expense
}

type Income struct {
	Owner  uuid.UUID `db:"owner"  json:"-"`
	Source string    `db:"source" json:"source"`
	Amount float64   `db:"amount" json:"amount"`
}

type Expense struct {
	BudgetID uuid.UUID `db:"budget_id" json:"-"`
	ID       uuid.UUID `db:"id"        json:"id"`
	Category string    `db:"category"  json:"category"`
	Amount   float64   `db:"amount"    json:"amount"`
	Fixed    bool      `db:"is_fixed"  json:"isFixed"`
	Slack    bool      `db:"is_slack"  json:"isSlack"`
}

func NewBudget(owner uuid.UUID, name string) *Budget {
	return &Budget{
		ID:       uuid.New(),
		Owner:    owner,
		Name:     name,
		Incomes:  []*Income{},
		Expenses: []*Expense{},
	}
}

func (b *Budget) SetBudgetIncome(source string, amount float64) {
	b.Incomes = append(b.Incomes, &Income{
		Owner:  b.ID,
		Source: source,
		Amount: amount,
	})
}

func (b *Budget) RemoveBudgetIncome(source string) {
	for i, income := range b.Incomes {
		if income.Source == source {
			b.Incomes = append(b.Incomes[:i], b.Incomes[i+1:]...)

			return
		}
	}
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
	b.Expenses = append(b.Expenses, &Expense{
		BudgetID: b.ID,
		Category: category,
		Amount:   amount,
		Fixed:    fixed,
		Slack:    slack,
	})
}

func (b *Budget) RemoveExpense(category string) {
	for i, expense := range b.Expenses {
		if expense.Category == category {
			b.Expenses = append(b.Expenses[:i], b.Expenses[i+1:]...)

			return
		}
	}
}
