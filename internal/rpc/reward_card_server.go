package rpc

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	yabav1 "yaba/gen/yaba/v1"
	"yaba/internal/database"
	"yaba/internal/model"
)

// RewardCardServer implements RewardCardService over the global reward card catalog.
type RewardCardServer struct {
	Pool *pgxpool.Pool
}

func (s *RewardCardServer) ListRewardCards(
	ctx context.Context,
	req *connect.Request[yabav1.ListRewardCardsRequest],
) (*connect.Response[yabav1.ListRewardCardsResponse], error) {
	msg := req.Msg

	var limit, offset *int

	if msg.Limit != nil {
		limit = new(int(msg.GetLimit()))
	}

	if msg.Offset != nil {
		offset = new(int(msg.GetOffset()))
	}

	cards, err := database.ListRewardCards(ctx, s.Pool,
		new(msg.GetIssuer()), new(msg.GetName()), new(msg.GetRegion()), limit, offset)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	out := make([]*yabav1.RewardCard, 0, len(cards))
	for _, card := range cards {
		out = append(out, rewardCardToProto(card))
	}

	return connect.NewResponse(&yabav1.ListRewardCardsResponse{RewardCards: out}), nil
}

func (s *RewardCardServer) CreateRewardCard(
	ctx context.Context,
	req *connect.Request[yabav1.CreateRewardCardRequest],
) (*connect.Response[yabav1.CreateRewardCardResponse], error) {
	msg := req.Msg
	card := &model.RewardCard{
		ID:         uuid.New(),
		Name:       msg.GetName(),
		Issuer:     msg.GetIssuer(),
		Region:     msg.GetRegion(),
		RewardType: msg.GetRewardType(),
	}

	for _, category := range msg.GetCategories() {
		card.RewardCategories = append(card.RewardCategories, &model.RewardCategory{
			CardID:   card.ID,
			Category: category.GetCategory(),
			Rate:     category.GetRate(),
		})
	}

	if err := database.CreateRewardCard(ctx, s.Pool, card); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&yabav1.CreateRewardCardResponse{RewardCard: rewardCardToProto(card)}), nil
}

func rewardCardToProto(card *model.RewardCard) *yabav1.RewardCard {
	out := &yabav1.RewardCard{
		Id:         card.ID.String(),
		Name:       card.Name,
		Issuer:     card.Issuer,
		Region:     card.Region,
		Version:    int32(card.Version), //nolint:gosec // versions are small
		RewardType: card.RewardType,
	}

	for _, category := range card.RewardCategories {
		out.Categories = append(out.Categories, &yabav1.RewardCategory{
			Category: category.Category,
			Rate:     category.Rate,
		})
	}

	return out
}
