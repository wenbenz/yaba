package importer

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"yaba/internal/budget"
	"yaba/internal/errors"

	"github.com/google/uuid"
)

type CsvExpenditureReader struct {
	header2index map[string]int
	owner        uuid.UUID
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

func (reader *CsvExpenditureReader) ReadRow(row []string) (*budget.Expenditure, error) {
	date, err := reader.getDate(row, "date")
	if err != nil {
		return nil, err
	}

	amount, err := reader.getFloat64(row, "amount")
	if err != nil {
		return nil, err
	}

	return &budget.Expenditure{
		Owner:          reader.owner,
		Name:           reader.getString(row, "name"),
		Date:           date,
		Amount:         amount,
		Method:         reader.getString(row, "method"),
		BudgetCategory: reader.getString(row, "budget_category"),
		RewardCategory: reader.getString(row, "reward_category"),
		Comment:        reader.getString(row, "comment"),
	}, nil
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

func ImportExpendituresFromCSVReader(owner uuid.UUID, r *csv.Reader) ([]*budget.Expenditure, error) {
	// First row is always headers.
	headers, err := r.Read()

	if err != nil {
		return nil, fmt.Errorf("received error reading headers: %w", err)
	}

	expenditureReader, err := NewCSVExpenditureReader(owner, headers)
	if err != nil {
		return nil, err
	}

	var expenditures []*budget.Expenditure

	for row, err := r.Read(); err != io.EOF; row, err = r.Read() {
		if err != nil {
			return nil, fmt.Errorf("unexpected error reading csv: %w", err)
		}

		// skip empty
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
