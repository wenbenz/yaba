package platform

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"yaba/internal/database"
	"yaba/internal/import"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func UploadSpendingsCSV(ctx context.Context, pool *pgxpool.Pool, user uuid.UUID, data io.Reader) error {
	csvReader := csv.NewReader(data)
	expenditures, err := importer.ImportExpendituresFromCSVReader(user, csvReader)

	if err != nil {
		return fmt.Errorf("failed to import: %w", err)
	}

	if err = database.PersistExpenditures(ctx, pool, expenditures); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	return nil
}
