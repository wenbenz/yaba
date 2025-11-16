package handlers

import (
	"encoding/csv"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"time"
	"yaba/internal/database"
)

func exportExpenditureHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		// Parse date parameters
		startDate, err := time.Parse("2006-01-02", r.URL.Query().Get("startDate"))
		if err != nil {
			http.Error(w, "invalid startDate format (use YYYY-MM-DD)", http.StatusBadRequest)

			return
		}

		endDate, err := time.Parse("2006-01-02", r.URL.Query().Get("endDate"))
		if err != nil {
			http.Error(w, "invalid endDate format (use YYYY-MM-DD)", http.StatusBadRequest)

			return
		}

		// Get transactions from database
		transactions, err := database.ListExpenditures(r.Context(), pool, nil, nil, nil, nil, startDate, endDate, nil, nil)
		if err != nil {
			http.Error(w, "failed to retrieve transactions", http.StatusInternalServerError)

			return
		}

		// Set headers for CSV download
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=expenditure.csv")

		// Create CSV writer
		csvWriter := csv.NewWriter(w)
		defer csvWriter.Flush()

		// Write header
		if err := csvWriter.Write([]string{
			"Date",
			"Name",
			"Amount",
			"Budget Category",
			"Reward Category",
			"Comment",
		}); err != nil {
			http.Error(w, "error writing CSV", http.StatusInternalServerError)

			return
		}

		// And update the corresponding data write:
		for _, t := range transactions {
			if err := csvWriter.Write([]string{
				t.Date.Format("2006-01-02"),
				t.Name,
				fmt.Sprintf("%.2f", t.Amount),
				t.BudgetCategory,
				t.RewardCategory,
				t.Comment,
			}); err != nil {
				http.Error(w, "error writing CSV", http.StatusInternalServerError)

				return
			}
		}
	}
}
