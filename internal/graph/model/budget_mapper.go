package model

import (
	"github.com/google/uuid"
	"yaba/internal/budget"
)

func BudgetFromNewBudgetInput(owner uuid.UUID, input *NewBudgetInput) *budget.Budget {
	budgetID := uuid.New()

	return &budget.Budget{
		ID:       budgetID,
		Owner:    owner,
		Name:     input.Name,
		Incomes:  incomesFromIncomeInput(budgetID, input.Incomes),
		Expenses: expensesFromExpenseInput(budgetID, input.Expenses),
	}
}

func BudgetFromUpdateBudgetInput(budgetID, owner uuid.UUID, input *UpdateBudgetInput) *budget.Budget {
	return &budget.Budget{
		ID:       budgetID,
		Owner:    owner,
		Name:     *input.Name,
		Incomes:  incomesFromIncomeInput(budgetID, input.Incomes),
		Expenses: expensesFromExpenseInput(budgetID, input.Expenses),
	}
}

func BudgetToBudgetResponse(b *budget.Budget) *BudgetResponse {
	return &BudgetResponse{
		ID:       b.ID.String(),
		Owner:    b.Owner.String(),
		Name:     b.Name,
		Incomes:  incomesToIncomeResponse(b.Incomes),
		Expenses: expensesToExpenseResponse(b.Expenses),
	}
}

func expensesToExpenseResponse(expenses []*budget.Expense) []*ExpenseResponse {
	ret := make([]*ExpenseResponse, len(expenses))
	for i, expense := range expenses {
		ret[i] = &ExpenseResponse{
			Category: expense.Category,
			Amount:   expense.Amount,
			IsFixed:  &expense.Fixed,
			IsSlack:  &expense.Slack,
		}
	}

	return ret
}

func incomesToIncomeResponse(incomes []*budget.Income) []*IncomeResponse {
	ret := make([]*IncomeResponse, len(incomes))
	for i, income := range incomes {
		ret[i] = &IncomeResponse{
			Source: income.Source,
			Amount: income.Amount,
		}
	}

	return ret
}

func expensesFromExpenseInput(budgetID uuid.UUID, input []*ExpenseInput) []*budget.Expense {
	expenses := make([]*budget.Expense, len(input))

	for i, expense := range input {
		expenses[i] = &budget.Expense{
			BudgetID: budgetID,
			Category: expense.Category,
			Amount:   expense.Amount,
		}

		if expense.IsFixed != nil {
			expenses[i].Fixed = *expense.IsFixed
		}

		if expense.IsSlack != nil {
			expenses[i].Slack = *expense.IsSlack
		}
	}

	return expenses
}

func incomesFromIncomeInput(budgetID uuid.UUID, input []*IncomeInput) []*budget.Income {
	incomes := make([]*budget.Income, len(input))

	for i, income := range input {
		incomes[i] = &budget.Income{
			Owner:  budgetID,
			Source: income.Source,
			Amount: income.Amount,
		}
	}

	return incomes
}
