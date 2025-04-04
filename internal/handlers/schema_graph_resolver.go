package handlers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.68

import (
	"context"
	"fmt"
	"time"
	"yaba/graph/model"
	"yaba/graph/server"
	"yaba/internal/ctxutil"
	"yaba/internal/database"

	"github.com/google/uuid"
)

// CreateBudget is the resolver for the createBudget field.
func (r *mutationResolver) CreateBudget(ctx context.Context, input model.NewBudgetInput) (*model.BudgetResponse, error) {
	user := ctxutil.GetUser(ctx)

	b, err := model.BudgetFromNewBudgetInput(user, &input)
	if err != nil {
		return nil, err
	}

	if err = database.PersistBudget(ctx, r.Pool, b); err != nil {
		return nil, err
	}

	return model.BudgetToBudgetResponse(b), nil
}

// UpdateBudget is the resolver for the updateBudget field.
func (r *mutationResolver) UpdateBudget(ctx context.Context, input model.UpdateBudgetInput) (*model.BudgetResponse, error) {
	user := ctxutil.GetUser(ctx)
	budgetID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid budget ID: %w", err)
	}

	// Check that user owns this budget; if not, this will fail.
	_, err = database.GetBudget(ctx, r.Pool, user, budgetID)
	if err != nil {
		return nil, fmt.Errorf("budget not found: %w", err)
	}

	// Persist budget
	b, err := model.BudgetFromUpdateBudgetInput(budgetID, user, &input)
	if err != nil {
		return nil, err
	}

	if err := database.PersistBudget(ctx, r.Pool, b); err != nil {
		return nil, err
	}

	return model.BudgetToBudgetResponse(b), nil
}

// CreateExpenditures is the resolver for the createExpenditures field.
func (r *mutationResolver) CreateExpenditures(ctx context.Context, input []*model.ExpenditureInput) (*bool, error) {
	user := ctxutil.GetUser(ctx)
	expenditures, err := model.ExpendituresFromExpenditureInput(user, input)

	var success bool

	if err != nil {
		return &success, err
	}

	if err = database.PersistExpenditures(ctx, r.Pool, expenditures); err != nil {
		return &success, err
	}

	success = true
	return &success, nil
}

// Budget is the resolver for the budget field.
func (r *queryResolver) Budget(ctx context.Context, id string) (*model.BudgetResponse, error) {
	user := ctxutil.GetUser(ctx)
	budgetID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid budget ID: %w", err)
	}

	b, err := database.GetBudget(ctx, r.Pool, user, budgetID)
	if err != nil {
		return nil, err
	}

	return model.BudgetToBudgetResponse(b), nil
}

// Budgets is the resolver for the budgets field.
func (r *queryResolver) Budgets(ctx context.Context, first *int) ([]*model.BudgetResponse, error) {
	limit := 10
	if first != nil {
		limit = *first
	}

	user := ctxutil.GetUser(ctx)
	b, err := database.GetBudgets(ctx, r.Pool, user, limit)
	if err != nil {
		return nil, err
	}

	out := make([]*model.BudgetResponse, len(b))
	for i := range b {
		out[i] = model.BudgetToBudgetResponse(b[i])
	}

	return out, nil
}

// Expenditures is the resolver for the expenditures field.
func (r *queryResolver) Expenditures(ctx context.Context, since *string, until *string, source *string, category *string, count *int, offset *int) ([]*model.ExpenditureResponse, error) {
	start := time.Unix(0, 0).Format(time.DateOnly)
	if since != nil {
		start = *since
	}

	end := time.Now().Format(time.DateOnly)
	if until != nil {
		end = *until
	}

	limit := 10
	if count != nil {
		limit = *count
	}

	sinceTime, err := time.Parse(time.DateOnly, start)
	if err != nil {
		return []*model.ExpenditureResponse{}, err
	}

	untilTime, err := time.Parse(time.DateOnly, end)
	if err != nil {
		return []*model.ExpenditureResponse{}, err
	}

	expenditures, err := database.ListExpenditures(ctx, r.Pool, sinceTime, untilTime, source, category, &limit, offset)
	if err != nil {
		return []*model.ExpenditureResponse{}, err
	}

	return model.ExpendituresToExpenitureResponse(expenditures), nil
}

// AggregatedExpenditures is the resolver for the aggregatedExpenditures field.
func (r *queryResolver) AggregatedExpenditures(ctx context.Context, since *string, until *string, span *model.Timespan, groupBy *model.GroupBy, aggregation *model.Aggregation) ([]*model.AggregatedExpendituresResponse, error) {
	var err error

	start := time.Unix(0, 0)
	if since != nil {
		start, err = time.Parse(time.DateOnly, *since)
		if err != nil {
			return []*model.AggregatedExpendituresResponse{}, fmt.Errorf("invalid start date: %w", err)
		}
	}

	end := time.Now()
	if until != nil {
		end, err = time.Parse(time.DateOnly, *until)
		if err != nil {
			return []*model.AggregatedExpendituresResponse{}, fmt.Errorf("invalid end date: %w", err)
		}
	}

	timespan := model.TimespanDay
	if span != nil {
		timespan = *span
	}

	gb := model.GroupByNone
	if groupBy != nil {
		gb = *groupBy
	}

	agg := model.AggregationSum
	if aggregation != nil {
		agg = *aggregation
	}

	aggregateExpenditures, err := database.AggregateExpenditures(ctx, r.Pool, start, end, timespan, agg, gb)
	if err != nil {
		return nil, fmt.Errorf("aggregatedExpenditures: %w", err)
	}

	return model.ExpenditureSummariesToAggregateExpenditures(aggregateExpenditures, timespan), nil
}

// Mutation returns server.MutationResolver implementation.
func (r *Resolver) Mutation() server.MutationResolver { return &mutationResolver{r} }

// Query returns server.QueryResolver implementation.
func (r *Resolver) Query() server.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
