package main

import (
	"encoding/csv"
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"time"
	"yaba/internal/test/helper"
)

func main() {
	user, _ := uuid.Parse("b49585ce-1ba0-4875-a99a-431b4c44c4d0")
	start, _ := time.Parse(time.DateOnly, "2025-01-01")
	expenditures := helper.MockExpenditures(1000, user, start, time.Now())

	f, _ := os.Create("generated.csv")
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println("Error closing file:", err)
		}
	}(f)

	csvWriter := csv.NewWriter(f)
	defer csvWriter.Flush()
	_ = csvWriter.Write([]string{"date", "amount", "name", "method", "budget_category", "reward_category", "comment"})

	for _, exp := range expenditures {
		err := csvWriter.Write([]string{exp.Date.Format(time.DateOnly), fmt.Sprintf("%.2f", exp.Amount),
			exp.Name, exp.Method, exp.BudgetCategory, exp.RewardCategory, exp.Comment})
		if err != nil {
			log.Println("Error writing to csv:", err)
		}
	}
}
