package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"
	"yaba/internal/model"
	"yaba/internal/test/helper"

	"github.com/google/uuid"
)

func main() {
	user, _ := uuid.Parse("b49585ce-1ba0-4875-a99a-431b4c44c4d0")
	start, _ := time.ParseInLocation(time.DateOnly, "2025-01-01", time.UTC)
	generator := helper.NewTestDataGenerator(user, 0)
	expenditures := generator.GenerateExpenditures(100, user, start, time.Now())

	f, _ := os.Create("generated.csv")
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println("Error closing file:", err)
		}
	}(f)

	csvWriter := csv.NewWriter(f)
	defer csvWriter.Flush()
	_ = csvWriter.Write(
		[]string{
			"date",
			"amount",
			"name",
			"method",
			"budget_category",
			"reward_category",
			"comment",
		},
	)

	id2method := map[uuid.UUID]*model.PaymentMethod{}
	for _, pm := range generator.PaymentMethods {
		id2method[pm.ID] = pm
	}

	for _, exp := range expenditures {
		err := csvWriter.Write(
			[]string{exp.Date.Format(time.DateOnly), fmt.Sprintf("%.2f", exp.Amount),
				exp.Name, id2method[exp.Method].DisplayName, exp.BudgetCategory, exp.RewardCategory, exp.Comment},
		)
		if err != nil {
			log.Println("Error writing to csv:", err)
		}
	}
}
