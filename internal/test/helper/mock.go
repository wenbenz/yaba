package helper

import (
	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
	"time"
	"yaba/internal/model"
)

func MockExpenditures(n int, owner uuid.UUID, startDate, endDate time.Time) []*model.Expenditure {
	expenditures := make([]*model.Expenditure, n)
	for i := range expenditures {
		expenditures[i] = &model.Expenditure{
			Owner:          owner,
			Name:           gofakeit.BeerStyle(),
			Amount:         gofakeit.Float64Range(0.01, 100),
			Date:           gofakeit.DateRange(startDate, endDate.AddDate(0, 0, 1)).Truncate(24 * time.Hour),
			Method:         gofakeit.CreditCardType(),
			BudgetCategory: gofakeit.FuelType(),
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
			CreatedTime: gofakeit.DateRange(startDate.UTC(), endDate.AddDate(0, 0, 1)),
			Source:      gofakeit.CarMaker() + ".csv",
		}
	}

	return expenditures
}

// func MakeMockExpenditureCSV() {
//	user, _ := uuid.Parse("b49585ce-1ba0-4875-a99a-431b4c44c4d0")
//	start, _ := time.Parse(time.DateOnly, "2025-01-01")
//	expenditures := MockExpenditures(1000, user, start, time.Now())
//	f, _ := os.Create("generated.csv")
//	defer f.Close()
//	csvWriter := csv.NewWriter(f)
//	defer csvWriter.Flush()
//	csvWriter.Write([]string{"date", "amount", "name", "method", "budget_category", "reward_category", "comment"})
//	for _, exp := range expenditures {
//		csvWriter.Write([]string{exp.Date.Format(time.DateOnly), fmt.Sprintf("%.2f", exp.Amount),
//			exp.Name, exp.Method, exp.BudgetCategory, exp.RewardCategory, exp.Comment})
//	}
//}
