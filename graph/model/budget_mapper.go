package model

import (
	"fmt"
	"github.com/google/uuid"
	"time"
	"yaba/internal/model"
)

func BudgetFromNewBudgetInput(owner uuid.UUID, input *NewBudgetInput) (*model.Budget, error) {
	budgetID := uuid.New()

	expenses, err := expensesFromExpenseInput(budgetID, input.Expenses)
	if err != nil {
		return nil, err
	}

	return &model.Budget{
		ID:       budgetID,
		Owner:    owner,
		Name:     input.Name,
		Incomes:  incomesFromIncomeInput(budgetID, input.Incomes),
		Expenses: expenses,
	}, nil
}

func BudgetFromUpdateBudgetInput(budgetID, owner uuid.UUID, input *UpdateBudgetInput) (*model.Budget, error) {
	expenses, err := expensesFromExpenseInput(budgetID, input.Expenses)
	if err != nil {
		return nil, err
	}

	return &model.Budget{
		ID:       budgetID,
		Owner:    owner,
		Name:     *input.Name,
		Incomes:  incomesFromIncomeInput(budgetID, input.Incomes),
		Expenses: expenses,
	}, nil
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

func ExpendituresFromExpenditureInput(user uuid.UUID, input []*ExpenditureInput) ([]*model.Expenditure, error) {
	expenditures := make([]*model.Expenditure, len(input))

	for i, expenditure := range input {
		date, err := time.ParseInLocation(time.DateOnly, expenditure.Date, time.UTC)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %w", err)
		}

		expenditures[i] = &model.Expenditure{
			Owner:          user,
			Date:           date,
			Amount:         expenditure.Amount,
			Name:           dereferenceOrEmpty(expenditure.Name),
			Method:         dereferenceOrEmpty(expenditure.Method),
			BudgetCategory: dereferenceOrEmpty(expenditure.BudgetCategory),
			RewardCategory: dereferenceOrEmpty(expenditure.RewardCategory),
			Comment:        dereferenceOrEmpty(expenditure.Comment),
			Source:         dereferenceOrEmpty(expenditure.Source),
		}
	}

	return expenditures, nil
}

func dereferenceOrEmpty(ptr *string) string {
	if ptr != nil {
		return *ptr
	}

	return ""
}

func expensesToExpenseResponse(expenses []*model.Expense) []*ExpenseResponse {
	ret := make([]*ExpenseResponse, len(expenses))

	for i, expense := range expenses {
		expenseID := expense.ID.String()
		ret[i] = &ExpenseResponse{
			ID:       &expenseID,
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

func expensesFromExpenseInput(budgetID uuid.UUID, input []*ExpenseInput) ([]*model.Expense, error) {
	expenses := make([]*model.Expense, len(input))

	for i, expense := range input {
		expenseID := uuid.New()

		var err error
		if expense.ID != nil {
			if expenseID, err = uuid.Parse(*expense.ID); err != nil {
				return nil, fmt.Errorf("failed to parse UUID: %w", err)
			}
		}

		expenses[i] = &model.Expense{
			ID:       expenseID,
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

	return expenses, nil
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
