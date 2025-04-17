package model

import (
	"yaba/internal/model"
)

// RewardCardToRewardCardResponse converts an internal reward card to a GraphQL response.
func RewardCardToRewardCardResponse(rc *model.RewardCard) *RewardCard {
	if rc == nil {
		return nil
	}

	return &RewardCard{
		ID:              rc.ID.String(),
		Name:            rc.Name,
		Issuer:          rc.Issuer,
		Region:          rc.Region,
		Version:         rc.Version,
		RewardRate:      rc.RewardRate,
		RewardType:      rc.RewardType,
		RewardCashValue: rc.RewardCashValue,
	}
}

// RewardCardFromRewardCardInput converts a GraphQL input to an internal reward card.
func RewardCardFromRewardCardInput(input RewardCardInput) *model.RewardCard {
	return &model.RewardCard{
		Name:            input.Name,
		Issuer:          input.Issuer,
		Region:          input.Region,
		RewardRate:      input.RewardRate,
		RewardType:      input.RewardType,
		RewardCashValue: input.RewardCashValue,
	}
}
