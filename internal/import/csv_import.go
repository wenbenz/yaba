package importer

import (
	"encoding/csv"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
	"io"
	"strconv"
	"strings"
	"time"
	"yaba/errors"
	"yaba/internal/database"
	"yaba/internal/model"
)

func UploadSpendingsCSV(ctx context.Context, pool *pgxpool.Pool, user uuid.UUID, data io.Reader, source string) error {
	csvReader := csv.NewReader(data)
	expenditures, err := ImportExpendituresFromCSVReader(user, csvReader)

	if err != nil {
		return fmt.Errorf("failed to import: %w", err)
	}

	for _, e := range expenditures {
		e.Source = source
	}

	if err = database.PersistExpenditures(ctx, pool, expenditures); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	return nil
}

func ImportExpendituresFromCSVReader(owner uuid.UUID, r *csv.Reader) ([]*model.Expenditure, error) {
	headers, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("received error reading headers: %w", err)
	}

	expenditureReader, err := NewCSVExpenditureReader(owner, headers)
	if err != nil {
		return nil, err
	}

	var expenditures []*model.Expenditure

	for row, err := r.Read(); err != io.EOF; row, err = r.Read() {
		if err != nil {
			return nil, fmt.Errorf("unexpected error reading csv: %w", err)
		}

		if len(row) == 0 {
			continue
		}

		expenditure, err := expenditureReader.ReadRow(row)
		if err != nil {
			return nil, err
		}

		expenditures = append(expenditures, expenditure)
	}

	return expenditures, err
}

type CsvExpenditureReader struct {
	header2index map[string]int
	owner        uuid.UUID
}

func NewCSVExpenditureReader(owner uuid.UUID, headers []string) (*CsvExpenditureReader, error) {
	reader := CsvExpenditureReader{
		owner:        owner,
		header2index: make(map[string]int),
	}

	if err := validateHeaders(headers); err != nil {
		return nil, err
	}

	for i, h := range headers {
		reader.header2index[h] = i
	}

	return &reader, nil
}

func (reader *CsvExpenditureReader) ReadRow(row []string) (*model.Expenditure, error) {
	date, err := reader.getDate(row, "date")
	if err != nil {
		return nil, err
	}

	amount, err := reader.getFloat64(row, "amount")
	if err != nil {
		return nil, err
	}

	rewardCategory := reader.getString(row, "reward_category")

	return &model.Expenditure{
		Owner:          reader.owner,
		Name:           reader.getString(row, "name"),
		Date:           date,
		Amount:         amount,
		Method:         reader.getString(row, "method"),
		BudgetCategory: reader.getString(row, "budget_category"),
		RewardCategory: strings.ToUpper(rewardCategory),
		Comment:        reader.getString(row, "comment"),
	}, nil
}

func (reader *CsvExpenditureReader) getString(row []string, key string) string {
	i, ok := reader.header2index[key]
	if !ok {
		return ""
	}

	return strings.ToLower(strings.TrimSpace(row[i]))
}

func (reader *CsvExpenditureReader) getDate(row []string, key string) (time.Time, error) {
	date, err := time.Parse(time.DateOnly, reader.getString(row, key))
	if err != nil {
		return time.Now(), fmt.Errorf("date must have format YYYY-MM-DD: %w", err)
	}

	return date, nil
}

func (reader *CsvExpenditureReader) getFloat64(row []string, key string) (float64, error) {
	s := reader.getString(row, key)
	s = strings.ReplaceAll(s, ",", " ")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "$", "")
	dollars, err := strconv.ParseFloat(s, 64)

	if err != nil {
		return 0, fmt.Errorf("failed to parse dollars from '%s': %w", s, err)
	}

	return dollars, nil
}

func validateHeaders(headers []string) error {
	allowedHeaders := []string{"date", "amount", "name", "method", "budget_category", "reward_category", "comment"}
	hasDate, hasAmount := false, false

	for _, h := range headers {
		switch h {
		case "date":
			hasDate = true
		case "amount":
			hasAmount = true
		default:
			allowed := false

			for _, allowedHeader := range allowedHeaders {
				if h == allowedHeader {
					allowed = true

					break
				}
			}

			if !allowed {
				return fmt.Errorf("unrecognized column '%s' in headers: %w", h, errors.InvalidInputError{Input: headers})
			}
		}
	}

	if !hasDate {
		return fmt.Errorf("missing required column 'date': %w", errors.InvalidInputError{Input: headers})
	}

	if !hasAmount {
		return fmt.Errorf("missing required column 'amount': %w", errors.InvalidInputError{Input: headers})
	}

	return nil
}
