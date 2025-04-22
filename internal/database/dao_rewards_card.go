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

	if err = setRewardCardCategories(ctx, pool, []*model.RewardCard{&card}); err != nil {
		return nil, err
	}

	return &card, nil
}

func ListRewardCards(ctx context.Context, pool *pgxpool.Pool,
	issuer, name, region *string, limit, offset *int) ([]*model.RewardCard, error) {
	cards, err := getCards(ctx, pool, issuer, name, region, limit, offset)
	if err != nil {
		return cards, err
	}

	if err = setRewardCardCategories(ctx, pool, cards); err != nil {
		return nil, err
	}

	return cards, nil
}

//nolint:cyclop
func getCards(ctx context.Context, pool *pgxpool.Pool,
	issuer *string, name *string, region *string, limit *int, offset *int) ([]*model.RewardCard, error) {
	query := squirrel.Select("*").
		From("rewards_card").
		OrderBy("name", "version DESC")

	if issuer != nil && *issuer != "" {
		query = query.Where(squirrel.ILike{"issuer": *issuer + "%"})
	}

	if name != nil && *name != "" {
		query = query.Where(squirrel.ILike{"name": *name + "%"})
	}

	if region != nil && *region != "" {
		query = query.Where(squirrel.ILike{"region": *region + "%"})
	}

	l := 10
	if limit != nil {
		l = *limit
	}

	query = query.Limit(uint64(l)) //nolint:gosec

	if offset != nil {
		query = query.Offset(uint64(*offset)) //nolint:gosec
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

func setRewardCardCategories(ctx context.Context, pool *pgxpool.Pool, cards []*model.RewardCard) error {
	if len(cards) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, len(cards))
	for i, card := range cards {
		ids[i] = card.ID
	}

	query, args, err := squirrel.Select("*").
		From("card_rewards").
		Where(squirrel.Eq{"card_id": ids}).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build reward categories query: %w", err)
	}

	var categories []*model.RewardCategory
	if err = pgxscan.Select(ctx, pool, &categories, query, args...); err != nil {
		return fmt.Errorf("failed to get reward categories: %w", err)
	}

	// Create a map for efficient lookup
	cardMap := make(map[uuid.UUID]*model.RewardCard)
	for _, card := range cards {
		cardMap[card.ID] = card
	}

	// Associate categories with their cards
	for _, category := range categories {
		if card, exists := cardMap[category.CardID]; exists {
			card.RewardCategories = append(card.RewardCategories, category)
		}
	}

	return nil
}

func CreateRewardCard(ctx context.Context, pool *pgxpool.Pool, reward *model.RewardCard) error {
	if err := validateRewardCard(reward); err != nil {
		return err
	}

	query, args, err := squirrel.Insert("rewards_card").
		Columns("id", "name", "region", "version", "issuer", "reward_type").
		Values(reward.ID, reward.Name, reward.Region, reward.Version, reward.Issuer,
			reward.RewardType).
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

	return nil
}
