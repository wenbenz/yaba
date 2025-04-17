package database

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"yaba/internal/ctxutil"
	"yaba/internal/model"
)

func ListPaymentMethods(ctx context.Context, pool *pgxpool.Pool) ([]*model.PaymentMethod, error) {
	query, args, err := squirrel.Select("*").
		From("payment_method").
		Where(squirrel.Eq{"owner": ctxutil.GetUser(ctx)}).
		OrderBy("display_name").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var methods []*model.PaymentMethod
	if err = pgxscan.Select(ctx, pool, &methods, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list payment methods: %w", err)
	}

	return methods, nil
}

func GetPaymentMethod(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*model.PaymentMethod, error) {
	pmQuery, pmArgs, err := squirrel.Select("*").
		From("payment_method").
		Where(squirrel.Eq{
			"id":    id,
			"owner": ctxutil.GetUser(ctx),
		}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build payment method query: %w", err)
	}

	var method model.PaymentMethod
	if err = pgxscan.Get(ctx, pool, &method, pmQuery, pmArgs...); err != nil {
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}

	// Second query: Get reward card if card_type is not nil
	if method.CardType != uuid.Nil {
		rcQuery, rcArgs, err := squirrel.Select("*").
			From("rewards_card").
			Where(squirrel.Eq{"id": method.CardType}).
			PlaceholderFormat(squirrel.Dollar).
			ToSql()
		if err != nil {
			return nil, fmt.Errorf("failed to build rewards card query: %w", err)
		}

		var reward model.RewardCard
		if err = pgxscan.Get(ctx, pool, &reward, rcQuery, rcArgs...); err != nil {
			return nil, fmt.Errorf("failed to get rewards card: %w", err)
		}

		method.Rewards = &reward
	}

	return &method, nil
}

func CreatePaymentMethod(ctx context.Context, pool *pgxpool.Pool, method *model.PaymentMethod) error {
	method.Owner = ctxutil.GetUser(ctx)

	query, args, err := squirrel.Insert("payment_method").
		Columns("id", "owner", "display_name", "card_type", "acquired_date", "cancel_by_date").
		Values(method.ID, method.Owner, method.DisplayName, method.CardType,
			method.AcquiredDate.UTC(), method.CancelByDate.UTC()).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build payment method query: %w", err)
	}

	if _, err = pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to create payment method: %w", err)
	}

	return nil
}

func UpdatePaymentMethod(ctx context.Context, pool *pgxpool.Pool, method *model.PaymentMethod) error {
	query, args, err := squirrel.Update("payment_method").
		Set("display_name", method.DisplayName).
		Set("acquired_date", method.AcquiredDate).
		Set("cancel_by_date", method.CancelByDate).
		Set("card_type", method.CardType).
		Where(squirrel.Eq{
			"id":    method.ID,
			"owner": ctxutil.GetUser(ctx),
		}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build payment method query: %w", err)
	}

	if _, err = pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to update payment method: %w", err)
	}

	return nil
}

func DeletePaymentMethod(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (bool, error) {
	query, args, err := squirrel.Delete("payment_method").
		Where(squirrel.Eq{
			"id":    id,
			"owner": ctxutil.GetUser(ctx),
		}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	tag, err := pool.Exec(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("failed to delete payment method: %w", err)
	}

	return tag.RowsAffected() > 0, nil
}
