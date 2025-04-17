package database_test

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"yaba/internal/database"
	"yaba/internal/model"
	"yaba/internal/test/helper"
)

func TestCreateRewardCard(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		reward      *model.RewardCard
		expectError bool
	}{
		{
			name: "valid reward card",
			reward: &model.RewardCard{
				ID:              uuid.New(),
				Name:            "Cash Back",
				Region:          "Canada",
				Version:         1,
				Issuer:          "Chase",
				RewardRate:      0.025,
				RewardType:      "cash",
				RewardCashValue: 0.025,
			},
		},
		{
			name: "valid reward card with different values",
			reward: &model.RewardCard{
				ID:              uuid.New(),
				Name:            "Travel Points",
				Region:          "Canada",
				Version:         1,
				Issuer:          "American Express",
				RewardRate:      0.03,
				RewardType:      "points",
				RewardCashValue: 0.01,
			},
		},
		{
			name:        "invalid reward card - nil",
			reward:      nil,
			expectError: true,
		},
		{
			name: "invalid reward card - missing name",
			reward: &model.RewardCard{
				ID:              uuid.New(),
				Region:          "USA",
				Version:         1,
				Issuer:          "Chase",
				RewardRate:      0.025,
				RewardType:      "cash",
				RewardCashValue: 0.025,
			},
			expectError: true,
		},
		{
			name: "invalid reward card - missing region",
			reward: &model.RewardCard{
				ID:              uuid.New(),
				Name:            "Cash Back",
				Version:         1,
				Issuer:          "Chase",
				RewardRate:      0.025,
				RewardType:      "cash",
				RewardCashValue: 0.025,
			},
			expectError: true,
		},
		{
			name: "invalid reward card - missing issuer",
			reward: &model.RewardCard{
				ID:              uuid.New(),
				Name:            "Cash Back",
				Region:          "Canada",
				Version:         1,
				RewardRate:      0.025,
				RewardType:      "cash",
				RewardCashValue: 0.025,
			},
			expectError: true,
		},
		{
			name: "invalid reward card - missing reward type",
			reward: &model.RewardCard{
				ID:              uuid.New(),
				Name:            "Cash Back",
				Region:          "Canada",
				Version:         1,
				Issuer:          "Chase",
				RewardRate:      0.025,
				RewardCashValue: 0.025,
			},
			expectError: true,
		},
		{
			name: "invalid reward card - zero reward rate",
			reward: &model.RewardCard{
				ID:              uuid.New(),
				Name:            "Cash Back",
				Region:          "Canada",
				Version:         1,
				Issuer:          "Chase",
				RewardType:      "cash",
				RewardCashValue: 0.025,
			},
			expectError: true,
		},
		{
			name: "invalid reward card - zero cash value",
			reward: &model.RewardCard{
				ID:         uuid.New(),
				Name:       "Cash Back",
				Region:     "Canada",
				Version:    1,
				Issuer:     "Chase",
				RewardRate: 0.025,
				RewardType: "cash",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pool := helper.GetTestPool()
			ctx := t.Context()

			err := database.CreateRewardCard(ctx, pool, tc.reward)
			if tc.expectError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			stored, err := database.GetRewardCard(ctx, pool, tc.reward.ID)
			require.NoError(t, err)

			require.Equal(t, tc.reward.ID, stored.ID)
			require.Equal(t, tc.reward.Name, stored.Name)
			require.Equal(t, tc.reward.Version, stored.Version)
			require.Equal(t, tc.reward.Issuer, stored.Issuer)
			require.Equal(t, tc.reward.RewardType, stored.RewardType)
			require.InEpsilon(t, tc.reward.RewardRate, stored.RewardRate, 0.0001)
			require.InEpsilon(t, tc.reward.RewardCashValue, stored.RewardCashValue, 0.0001)
		})
	}
}

func TestCreateRewardCardVersioning(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	ctx := t.Context()

	card1 := &model.RewardCard{
		ID:              uuid.New(),
		Name:            "Freedom Flex",
		Version:         1,
		Issuer:          "Chase",
		Region:          "USA",
		RewardRate:      0.05,
		RewardType:      "cash",
		RewardCashValue: 0.05,
	}

	card2 := &model.RewardCard{
		ID:              uuid.New(),
		Name:            "Freedom Flex",
		Version:         2,
		Issuer:          "Chase",
		Region:          "USA",
		RewardRate:      0.03,
		RewardType:      "cash",
		RewardCashValue: 0.03,
	}

	// Create first card
	err := database.CreateRewardCard(ctx, pool, card1)
	require.NoError(t, err)

	// Create second card with same name
	err = database.CreateRewardCard(ctx, pool, card2)
	require.NoError(t, err)

	// Verify first card
	stored1, err := database.GetRewardCard(ctx, pool, card1.ID)
	require.NoError(t, err)
	require.Equal(t, card1.Name, stored1.Name)
	require.Equal(t, 1, stored1.Version)

	// Verify second card
	stored2, err := database.GetRewardCard(ctx, pool, card2.ID)
	require.NoError(t, err)
	require.Equal(t, card2.Name, stored2.Name)
	require.Equal(t, 2, stored2.Version)

	// Verify latest by name returns the newer version
	latest, err := database.GetLatestRewardCardByName(ctx, pool, card1.Name)
	require.NoError(t, err)
	require.Equal(t, card2.ID, latest.ID)
	require.Equal(t, 2, latest.Version)
}
