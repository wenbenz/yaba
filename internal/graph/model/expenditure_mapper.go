package model

import (
	"fmt"
	"strconv"
	"time"

	"yaba/internal/budget"
)

func ExpendituresToExpenitureResponse(expenditures []*budget.Expenditure) []*ExpenditureResponse {
	ret := make([]*ExpenditureResponse, len(expenditures))

	for i, obj := range expenditures {
		id := strconv.Itoa(obj.ID)
		owner := obj.Owner.String()
		amount := fmt.Sprintf("%.2f", obj.Amount)
		date := obj.Date.Format(time.DateOnly)
		cat := obj.RewardCategory.String
		created := obj.CreatedTime.Format(time.RFC3339)

		ret[i] = &ExpenditureResponse{
			ID:             &id,
			Owner:          &owner,
			Name:           &obj.Name,
			Amount:         &amount,
			Date:           &date,
			Method:         &obj.Method,
			BudgetCategory: &obj.BudgetCategory,
			RewardCategory: &cat,
			Comment:        &obj.Comment,
			Created:        &created,
			Source:         &obj.Source,
		}
	}

	return ret
}
