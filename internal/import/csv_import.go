package importer

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"yaba/internal/budget"

	"github.com/google/uuid"
)

type csvExpenditureReader struct {
	header2index map[string]int
	owner uuid.UUID
}

func (reader *csvExpenditureReader) getString(row []string, key string) string {
	i, ok := reader.header2index[key]
	if !ok {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(row[i]))
}

func (reader *csvExpenditureReader) getDate(row []string, key string) (time.Time, error) {
	date, err := time.Parse(time.DateOnly, reader.getString(row, key))
	if err != nil {
		return time.Now(), fmt.Errorf("date must have format YYYY-MM-DD: %w", err)
	}
	return date, nil
}

func (reader *csvExpenditureReader) getCents(row []string, key string) (int, error) {
	s := reader.getString(row, key)
	dollarString := s
	cents := 0
	var err error
	// If string is at least 3 characters and the third-last character is a ',' or '.'
	// then we have a cents component. e.g. 1.23 and 1,23 are both valid.
	if len(s) > 2 && (s[len(s)-3] == '.' || s[len(s)-3] == ',') {
		dollarString = s[:len(s) - 3]
		if cents, err = strconv.Atoi(s[len(s)-2:]); err != nil {
			return 0, fmt.Errorf("failed to parse cents from '%s': %w", s, err)
		}
	}
	dollarString = strings.ReplaceAll(dollarString, ",", " ")
	dollarString = strings.ReplaceAll(dollarString, " ", "")
	
	dollars, err := strconv.Atoi(dollarString)
	if err != nil {
		return 0, fmt.Errorf("failed to parse dollars from '%s': %w", s, err)
	}

	return dollars * 100 + cents, nil
}

func (reader *csvExpenditureReader) readRow(row[]string) (*budget.Expenditure, error) {
	date, err := reader.getDate(row, "date")
	if err != nil {
		return nil, err
	}

	amount, err := reader.getCents(row, "amount")
	if err != nil {
		return nil, err
	}

	return &budget.Expenditure{
		Owner: reader.owner,
		Name: reader.getString(row, "name"),
		Date: date,
		Amount: amount,
		Method: reader.getString(row, "method"),
		BudgetCategory: reader.getString(row, "budget_category"),
		RewardCategory: reader.getString(row, "reward_category"),
		Comment: reader.getString(row, "comment"),
	}, nil
}

func newCSVExpenditureReader(owner uuid.UUID, headers []string) (*csvExpenditureReader, error) {
	reader := csvExpenditureReader{
		owner: owner,
		header2index: make(map[string]int),
	}
	
	hasDate, hasAmount := false, false
	for i, h := range headers {
		if h == "date" {
			hasDate = true
		} else if h == "amount" {
			hasAmount = true
		} else if h != "name" && h != "method" && h != "budget_category" && h != "reward_category" && h != "comment" {
			return nil, fmt.Errorf("unrecognized column '%s'", h)
		}

		reader.header2index[h] = i
	}

	if !hasDate {
		return nil, fmt.Errorf("missing required column 'date'")
	}

	if !hasAmount {
		return nil, fmt.Errorf("missing required column 'amount'")
	}

	return &reader, nil
}

func ImportExpendituresFromCSVReader(owner uuid.UUID, r *csv.Reader) ([]*budget.Expenditure, error) {
	// First row is always headers.
	headers, err := r.Read()

	if err != nil {
		return nil, fmt.Errorf("received error reading headers: %w", err)
	}

	expenditureReader, err := newCSVExpenditureReader(owner, headers)
	if err != nil {
		return nil, err
	}

	var expenditures []*budget.Expenditure
	for row, err := r.Read(); err != io.EOF; row, err = r.Read(){
		if err != nil {
			return nil, err
		}
		
		// skip empty
		if len(row) == 0 {
			continue;
		}

		expenditure, err := expenditureReader.readRow(row)
		if err != nil {
			return nil, err
		}

		expenditures = append(expenditures, expenditure)
	}

	return expenditures, err
}
