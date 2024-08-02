package platform

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"yaba/internal/database"
	"yaba/internal/errors"
	"yaba/internal/import"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func UploadSpendingsCSV(ctx context.Context, pool *pgxpool.Pool, data io.Reader) error {
	u := ctx.Value("user")
	if u == nil {
		return errors.InvalidStateError{Message: "no user in context"}
	}
	owner, err := uuid.Parse(u.(string))
	if err != nil {
		return fmt.Errorf("failed to parse user context: %w", err)
	}

	csvReader := csv.NewReader(data)
	expenditures, err := importer.ImportExpendituresFromCSVReader(owner, csvReader)

	if err != nil {
		return fmt.Errorf("failed to import: %w", err)
	}

	if err = database.PersistExpenditures(ctx, pool, expenditures); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	return nil
}
