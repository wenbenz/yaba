package database

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"yaba/graph/model"
	"yaba/internal/ctxutil"

	"yaba/internal/budget"

	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const listExpenditures = `
SELECT * FROM expenditure
WHERE owner = $1
  AND date >= $2
  AND date <= $3
ORDER BY date, id
LIMIT $4;
`

const insertExpenditure = `
INSERT INTO expenditure (owner, name, amount, date, method, budget_category, reward_category, comment, created, source)
VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, '')::reward_category, $8, NOW(), $9)
`

func ListExpenditures(ctx context.Context, pool *pgxpool.Pool, since, until time.Time, limit int) ([]*budget.Expenditure, error) {
	var expenditures []*budget.Expenditure
	if err := pgxscan.Select(ctx, pool, &expenditures, listExpenditures, ctxutil.GetUser(ctx), since, until, limit); err != nil {
		return nil, fmt.Errorf("failed to get expenditures: %w", err)
	}

	return expenditures, nil
}

func AggregateExpenditures(ctx context.Context, pool *pgxpool.Pool, startDate, endDate time.Time,
	timespan model.Timespan, aggregation model.Aggregation, groupBy model.GroupBy) ([]*budget.ExpenditureSummary, error) {

	var category string
	switch groupBy {
	case model.GroupByNone:
		category = "'Total'"
	case model.GroupByBudgetCategory:
		category = "budget_category"
	case model.GroupByRewardCategory:
		category = "reward_category"
	}

	date := "date"
	if timespan != model.TimespanDay {
		// This group-by will cause postgres to do a sequential scan if the timespan is not "DAY".
		// We can fix this with a functional index, but then the date column becomes immutable.
		// We can also group by date and aggregate the month/year in code to improve performance
		// the selectivity becomes low (due to more/older data).
		date = fmt.Sprintf("date_trunc('%s', date)", timespan.String())
	}

	sq := squirrel.Select(date+" as date",
		category+" as category",
		fmt.Sprintf("%s(amount) as amount", aggregation.String())).
		From("expenditure").
		Where("owner = $1 AND date >= $2 AND date <= $3", ctxutil.GetUser(ctx), startDate, endDate).
		GroupBy(date).
		OrderBy("date ASC")

	if groupBy != model.GroupByNone {
		sq.GroupBy(strings.ToLower(groupBy.String()))
	}

	query, args, err := sq.ToSql()
	if err != nil {
		return []*budget.ExpenditureSummary{}, fmt.Errorf("failed to build query: %w", err)
	}

	var expenditures []*budget.ExpenditureSummary
	err = pgxscan.Select(ctx, pool, &expenditures, query, args...)

	if err != nil {
		log.Println("Error executing query: ", query, args, err)
		return []*budget.ExpenditureSummary{}, fmt.Errorf("failed to get expenditures: %w", err)
	}

	return expenditures, nil
}

func PersistExpenditures(ctx context.Context, pool *pgxpool.Pool, expenditures []*budget.Expenditure,
) error {
	batch := &pgx.Batch{}
	for _, e := range expenditures {
		batch.Queue(insertExpenditure, e.Owner, e.Name, e.Amount, e.Date,
			e.Method, e.BudgetCategory, e.RewardCategory, e.Comment, e.Source)
	}

	if err := pool.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("failed to save batch of expenditures: %w", err)
	}

	return nil
}
