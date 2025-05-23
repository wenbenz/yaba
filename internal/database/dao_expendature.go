package database

import (
	"context"
	"fmt"
	"strings"
	"time"
	"yaba/internal/ctxutil"
	"yaba/internal/model"

	"github.com/google/uuid"

	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ListExpenditures(
	ctx context.Context,
	pool *pgxpool.Pool,
	filter, category, paymentMethod, source *string,
	since, until time.Time,
	limit, cursor *int,
) ([]*model.Expenditure, error) {
	sq := squirrel.Select("*").
		From("expenditure").
		Where(`owner = ? AND date >= ? AND date <= ?`, ctxutil.GetUser(ctx), since.UTC(), until.UTC()).
		OrderBy("date DESC, id DESC").
		PlaceholderFormat(squirrel.Dollar)

	if filter != nil {
		sq = sq.Where(squirrel.Or{
			squirrel.ILike{"name": "%" + *filter + "%"},
			squirrel.ILike{"comment": "%" + *filter + "%"},
		})
	}

	if category != nil {
		sq = sq.Where(squirrel.Eq{"budget_category": *category})
	}

	if paymentMethod != nil {
		sq = sq.Where(squirrel.Eq{"method": *paymentMethod})
	}

	if source != nil {
		sq = sq.Where(squirrel.Eq{"source": *source})
	}

	if limit != nil {
		sq = sq.Limit(uint64(*limit)) //nolint:gosec
	}

	if cursor != nil {
		sq = sq.Offset(uint64(*cursor)) //nolint:gosec
	}

	var expenditures []*model.Expenditure
	var query string
	var args []interface{}
	var err error

	if query, args, err = sq.ToSql(); err == nil {
		err = pgxscan.Select(ctx, pool, &expenditures, query, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get expenditures: %w", err)
	}

	return expenditures, nil
}

func AggregateExpenditures(
	ctx context.Context,
	pool *pgxpool.Pool,
	startDate, endDate time.Time,
	timespan model.Timespan,
	aggregation model.Aggregation,
	groupBy model.GroupBy,
) ([]*model.ExpenditureSummary, error) {
	var category string
	var categoryDefault string

	switch groupBy {
	case model.GroupByNone:
		category = "'Total'"
	case model.GroupByBudgetCategory:
		category = "expense_id"
		categoryDefault = uuid.Nil.String()
	case model.GroupByRewardCategory:
		category = "reward_category"
	}

	date := "date"
	if timespan != model.TimespanDay {
		// This group-by will cause postgres to do a sequential scan if the timespan is not "DAY".
		// We can fix this with a functional index, but then the date column becomes immutable.
		// We can also group by date and aggregate the month/year in code to improve performance
		// the selectivity becomes low (due to more/older data).
		date = fmt.Sprintf("date_trunc('%s', date)", timespan)
	}

	sq := squirrel.Select(date+" as date",
		fmt.Sprintf("COALESCE(%s::text, '%s') as category", category, categoryDefault),
		string(aggregation)+"(amount) as amount").
		From("expenditure").
		Where("owner = $1 AND date >= $2 AND date <= $3", ctxutil.GetUser(ctx), startDate, endDate).
		GroupBy(date).
		OrderBy("date ASC")

	if groupBy != model.GroupByNone {
		sq = sq.GroupBy(category)
		sq = sq.OrderBy(category)
	}

	query, args, err := sq.ToSql()
	if err != nil {
		return []*model.ExpenditureSummary{}, fmt.Errorf("failed to build query: %w", err)
	}

	var expenditures []*model.ExpenditureSummary
	err = pgxscan.Select(ctx, pool, &expenditures, query, args...)

	if err != nil {
		return []*model.ExpenditureSummary{}, fmt.Errorf("failed to get expenditures: %w", err)
	}

	// set everything to utc
	for _, e := range expenditures {
		e.StartDate = e.StartDate.UTC()
	}

	return expenditures, nil
}

func PersistExpenditures(
	ctx context.Context,
	pool *pgxpool.Pool,
	expenditures []*model.Expenditure,
) error {
	// If budget exists, map the expense ID to the expenditure's expense_id
	budgets, err := GetBudgets(ctx, pool, ctxutil.GetUser(ctx), 1)
	if err != nil {
		return err
	}

	budgetMap := make(map[string]uuid.UUID)

	for _, budget := range budgets {
		for _, expense := range budget.Expenses {
			budgetMap[strings.ToLower(expense.Category)] = expense.ID
		}
	}

	for _, expenditure := range expenditures {
		if expenditure.BudgetCategory != "" && expenditure.ExpenseID == uuid.Nil {
			expenditure.ExpenseID = budgetMap[strings.ToLower(expenditure.BudgetCategory)]
		}
	}

	batch := &pgx.Batch{}

	for _, e := range expenditures {
		query, args, err := squirrel.Insert("expenditure").
			Columns("owner", "name", "amount", "date",
				"method", "budget_category", "reward_category", "comment", "source", "expense_id").
			Values(e.Owner, e.Name, e.Amount, e.Date,
				e.Method, e.BudgetCategory, e.RewardCategory, e.Comment, e.Source, e.ExpenseID).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build query: %w", err)
		}

		batch.Queue(query, args...)
	}

	if err = pool.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("failed to save batch of expenditures: %w", err)
	}

	return nil
}

func ClassifyExpendituresWithNewCategory(
	ctx context.Context,
	batch *pgx.Batch,
	category string,
	expenseID uuid.UUID,
) error {
	query, args, err := squirrel.Update("expenditure").
		Where(map[string]interface{}{
			"owner":           ctxutil.GetUser(ctx),
			"budget_category": category,
			"expense_id":      uuid.Nil,
		}).
		Set("expense_id", expenseID).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	batch.Queue(query, args...)

	return nil
}
