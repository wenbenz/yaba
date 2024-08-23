package helper

import (
	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
	"time"
	"yaba/internal/budget"
)

func MockExpenditures(n int, owner uuid.UUID, startDate, endDate time.Time) []*budget.Expenditure {
	expenditures := make([]*budget.Expenditure, n)
	for i := range expenditures {
		expenditures[i] = &budget.Expenditure{
			Owner:          owner,
			Name:           gofakeit.Company(),
			Amount:         gofakeit.Float64Range(0, 1000), //nolint:mnd
			Date:           gofakeit.DateRange(startDate, endDate.AddDate(0, 0, 1)),
			Method:         gofakeit.CreditCardType(),
			BudgetCategory: gofakeit.BeerStyle(),
			RewardCategory: gofakeit.RandString([]string{
				"DRUG_STORE",
				"ENTERTAINMENT",
				"FURNITURE",
				"GAS",
				"GROCERY",
				"HOME_IMPROVEMENT",
				"HOTEL",
				"PUBLIC_TRANSPORTATION",
				"RECURRING_BILL",
				"RESTAURANT",
			}),
			Comment:     gofakeit.HipsterSentence(5), //nolint:mnd
			CreatedTime: gofakeit.DateRange(startDate, endDate.AddDate(0, 0, 1)),
			Source:      gofakeit.Word() + ".csv",
		}
	}

	return expenditures
}
