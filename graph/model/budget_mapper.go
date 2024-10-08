package model

import (
	"github.com/google/uuid"
	"yaba/internal/model"
)

func BudgetFromNewBudgetInput(owner uuid.UUID, input *NewBudgetInput) *model.Budget {
	budgetID := uuid.New()

	return &model.Budget{
		ID:       budgetID,
		Owner:    owner,
		Name:     input.Name,
		Incomes:  incomesFromIncomeInput(budgetID, input.Incomes),
		Expenses: expensesFromExpenseInput(budgetID, input.Expenses),
	}
}

func BudgetFromUpdateBudgetInput(budgetID, owner uuid.UUID, input *UpdateBudgetInput) *model.Budget {
	return &model.Budget{
		ID:       budgetID,
		Owner:    owner,
		Name:     *input.Name,
		Incomes:  incomesFromIncomeInput(budgetID, input.Incomes),
		Expenses: expensesFromExpenseInput(budgetID, input.Expenses),
	}
}

func BudgetToBudgetResponse(b *model.Budget) *BudgetResponse {
	id, owner, name := b.ID.String(), b.Owner.String(), b.Name

	return &BudgetResponse{
		ID:       &id,
		Owner:    &owner,
		Name:     &name,
		Incomes:  incomesToIncomeResponse(b.Incomes),
		Expenses: expensesToExpenseResponse(b.Expenses),
	}
}

func expensesToExpenseResponse(expenses []*model.Expense) []*ExpenseResponse {
	ret := make([]*ExpenseResponse, len(expenses))
	for i, expense := range expenses {
		ret[i] = &ExpenseResponse{
			Category: &expense.Category,
			Amount:   &expense.Amount,
			IsFixed:  &expense.Fixed,
			IsSlack:  &expense.Slack,
		}
	}

	return ret
}

func incomesToIncomeResponse(incomes []*model.Income) []*IncomeResponse {
	ret := make([]*IncomeResponse, len(incomes))
	for i, income := range incomes {
		ret[i] = &IncomeResponse{
			Source: &income.Source,
			Amount: &income.Amount,
		}
	}

	return ret
}

func expensesFromExpenseInput(budgetID uuid.UUID, input []*ExpenseInput) []*model.Expense {
	expenses := make([]*model.Expense, len(input))

	for i, expense := range input {
		expenses[i] = &model.Expense{
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

func incomesFromIncomeInput(budgetID uuid.UUID, input []*IncomeInput) []*model.Income {
	incomes := make([]*model.Income, len(input))

	for i, income := range input {
		incomes[i] = &model.Income{
			Owner:  budgetID,
			Source: income.Source,
			Amount: income.Amount,
		}
	}

	return incomes
}
