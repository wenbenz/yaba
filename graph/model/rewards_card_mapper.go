package model

import (
	"yaba/internal/model"
)

// RewardCardToRewardCardResponse converts an internal reward card to a GraphQL response.
func RewardCardToRewardCardResponse(rc *model.RewardCard) *RewardCard {
	if rc == nil {
		return nil
	}

	card := &RewardCard{
		ID:         rc.ID.String(),
		Name:       rc.Name,
		Issuer:     rc.Issuer,
		Region:     rc.Region,
		Version:    rc.Version,
		RewardType: rc.RewardType,
	}

	for _, category := range rc.RewardCategories {
		card.Categories = append(card.Categories, &RewardCategory{
			Category: category.Category,
			Rate:     category.Rate,
		})
	}

	return card
}

// RewardCardFromRewardCardInput converts a GraphQL input to an internal reward card.
func RewardCardFromRewardCardInput(input RewardCardInput) *model.RewardCard {
	card := &model.RewardCard{
		Name:       input.Name,
		Issuer:     input.Issuer,
		Region:     input.Region,
		RewardType: input.RewardType,
	}

	for _, category := range input.RewardCategories {
		card.RewardCategories = append(card.RewardCategories, &model.RewardCategory{
			Category: category.Category,
			Rate:     category.Rate,
		})
	}

	return card
}
