package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"yaba/internal/budget"
	"yaba/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

func UploadBudget(ctx context.Context, pool *pgxpool.Pool, data io.Reader) error {
	var budget *budget.ExternalBudget
	if err := json.NewDecoder(data).Decode(&budget); err != nil {
		return fmt.Errorf("failed to decode budget: %w", err)
	}

	if err := database.PersistBudget(ctx, pool, budget.ToInternal()); err != nil {
		return fmt.Errorf("failed to persist budget: %w", err)
	}

	return nil
}
