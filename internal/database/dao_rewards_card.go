package database

import (
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
	"yaba/internal/model"
)

var (
	ErrNilRewardCard     = errors.New("reward card is nil")
	ErrMissingID         = errors.New("reward card ID is required")
	ErrMissingName       = errors.New("reward card name is required")
	ErrMissingRegion     = errors.New("reward card region is required")
	ErrMissingIssuer     = errors.New("reward card issuer is required")
	ErrMissingRewardType = errors.New("reward type is required")
	ErrMissingRewardRate = errors.New("reward rate is required")
	ErrMissingCashValue  = errors.New("reward cash value is required")
)

func GetRewardCard(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*model.RewardCard, error) {
	query, args, err := squirrel.Select("*").
		From("rewards_card").
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var card model.RewardCard
	if err = pgxscan.Get(ctx, pool, &card, query, args...); err != nil {
		return nil, fmt.Errorf("failed to get reward card: %w", err)
	}

	return &card, nil
}

func ListRewardCards(ctx context.Context, pool *pgxpool.Pool,
	issuer, name, region *string) ([]*model.RewardCard, error) {
	query := squirrel.Select("*").
		From("rewards_card").
		OrderBy("name", "version DESC")

	if issuer != nil && *issuer != "" {
		query = query.Where(squirrel.Eq{"issuer": *issuer})
	}

	if name != nil && *name != "" {
		query = query.Where(squirrel.Eq{"name": *name})
	}

	if region != nil && *region != "" {
		query = query.Where(squirrel.Eq{"region": *region})
	}

	sql, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var cards []*model.RewardCard
	if err = pgxscan.Select(ctx, pool, &cards, sql, args...); err != nil {
		return nil, fmt.Errorf("failed to list reward cards: %w", err)
	}

	return cards, nil
}

func CreateRewardCard(ctx context.Context, pool *pgxpool.Pool, reward *model.RewardCard) error {
	if err := validateRewardCard(reward); err != nil {
		return err
	}

	query, args, err := squirrel.Insert("rewards_card").
		Columns("id", "name", "region", "version", "issuer", "reward_type", "reward_rate", "reward_cash_value").
		Values(reward.ID, reward.Name, reward.Region, reward.Version, reward.Issuer,
			reward.RewardType, reward.RewardRate, reward.RewardCashValue).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build rewards card query: %w", err)
	}

	if _, err = pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to create rewards card: %w", err)
	}

	return nil
}

func validateRewardCard(reward *model.RewardCard) error {
	if reward == nil {
		return ErrNilRewardCard
	}

	if reward.ID == uuid.Nil {
		return ErrMissingID
	}

	if reward.Name == "" {
		return ErrMissingName
	}

	if reward.Region == "" {
		return ErrMissingRegion
	}

	if reward.Issuer == "" {
		return ErrMissingIssuer
	}

	if reward.RewardType == "" {
		return ErrMissingRewardType
	}

	if reward.RewardRate == 0 {
		return ErrMissingRewardRate
	}

	if reward.RewardCashValue == 0 {
		return ErrMissingCashValue
	}

	return nil
}
