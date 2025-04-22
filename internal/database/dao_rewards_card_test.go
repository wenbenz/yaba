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
}

//nolint:cyclop,paralleltest,tparallel
func TestListRewardCards(t *testing.T) {
	pool := helper.GetTestPool()
	ctx := t.Context()

	chase := "Chase"
	amex := "Amex"
	td := "TD"
	cashBackPlus := "Cash Back Plus"
	travelPoints := "Travel Points"
	canada := "Canada"
	usa := "USA"

	// Create test data with more varied combinations
	cards := []*model.RewardCard{
		{
			ID:              uuid.New(),
			Name:            cashBackPlus,
			Region:          canada,
			Version:         1,
			Issuer:          chase,
			RewardRate:      0.025,
			RewardType:      "cash",
			RewardCashValue: 0.025,
		},
		{
			ID:              uuid.New(),
			Name:            travelPoints,
			Region:          usa,
			Version:         1,
			Issuer:          amex,
			RewardRate:      0.03,
			RewardType:      "points",
			RewardCashValue: 0.01,
		},
		{
			ID:              uuid.New(),
			Name:            "Cash Back Basic",
			Region:          canada,
			Version:         1,
			Issuer:          td,
			RewardRate:      0.01,
			RewardType:      "cash",
			RewardCashValue: 0.01,
		},
		{
			ID:              uuid.New(),
			Name:            "Premium Card",
			Region:          usa,
			Version:         1,
			Issuer:          chase,
			RewardRate:      0.02,
			RewardType:      "cash",
			RewardCashValue: 0.02,
		},
	}

	// Clear the table before inserting test data
	_, _ = pool.Exec(t.Context(), "TRUNCATE TABLE rewards_card")

	// Insert test data
	for _, card := range cards {
		err := database.CreateRewardCard(ctx, pool, card)
		require.NoError(t, err)
	}

	tests := []struct {
		name         string
		issuer       *string
		cardName     *string
		region       *string
		expectedLen  int
		checkResults func([]*model.RewardCard) bool
	}{
		{
			name:         "no filters",
			expectedLen:  4,
			checkResults: func(_ []*model.RewardCard) bool { return true },
		},
		{
			name:        "issuer only - Chase",
			issuer:      &chase,
			expectedLen: 2,
			checkResults: func(cards []*model.RewardCard) bool {
				for _, c := range cards {
					if c.Issuer != chase {
						return false
					}
				}

				return true
			},
		},
		{
			name:        "name only",
			cardName:    &cashBackPlus,
			expectedLen: 1,
			checkResults: func(cards []*model.RewardCard) bool {
				return len(cards) == 1 && cards[0].Name == cashBackPlus
			},
		},
		{
			name:        "region only - Canada",
			region:      &canada,
			expectedLen: 2,
			checkResults: func(cards []*model.RewardCard) bool {
				for _, c := range cards {
					if c.Region != canada {
						return false
					}
				}

				return true
			},
		},
		{
			name:        "issuer only - TD",
			issuer:      &td,
			expectedLen: 1,
			checkResults: func(cards []*model.RewardCard) bool {
				return len(cards) == 1 && cards[0].Issuer == td
			},
		},
		{
			name:        "region only - USA",
			region:      &usa,
			expectedLen: 2,
			checkResults: func(cards []*model.RewardCard) bool {
				for _, c := range cards {
					if c.Region != usa {
						return false
					}
				}

				return true
			},
		},
		{
			name:        "issuer and name",
			issuer:      &amex,
			cardName:    &travelPoints,
			expectedLen: 1,
			checkResults: func(cards []*model.RewardCard) bool {
				return len(cards) == 1 && cards[0].Issuer == amex && cards[0].Name == travelPoints
			},
		},
		{
			name:        "issuer and region",
			issuer:      &chase,
			region:      &usa,
			expectedLen: 1,
			checkResults: func(cards []*model.RewardCard) bool {
				return len(cards) == 1 && cards[0].Issuer == chase && cards[0].Region == usa
			},
		},
		{
			name:        "name and region",
			cardName:    &travelPoints,
			region:      &usa,
			expectedLen: 1,
			checkResults: func(cards []*model.RewardCard) bool {
				return len(cards) == 1 && cards[0].Name == travelPoints && cards[0].Region == usa
			},
		},
		{
			name:        "all filters",
			issuer:      &chase,
			cardName:    &cashBackPlus,
			region:      &canada,
			expectedLen: 1,
			checkResults: func(cards []*model.RewardCard) bool {
				return len(cards) == 1 &&
					cards[0].Issuer == chase &&
					cards[0].Name == cashBackPlus &&
					cards[0].Region == canada
			},
		},
		{
			name:         "no matches - wrong issuer",
			issuer:       ptr("NonExistent"),
			expectedLen:  0,
			checkResults: func(cards []*model.RewardCard) bool { return len(cards) == 0 },
		},
		{
			name:         "no matches - wrong combination",
			issuer:       &amex,
			region:       &canada,
			expectedLen:  0,
			checkResults: func(cards []*model.RewardCard) bool { return len(cards) == 0 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cards, err := database.ListRewardCards(ctx, pool, tt.issuer, tt.cardName, tt.region)
			require.NoError(t, err)
			require.Len(t, cards, tt.expectedLen)
			require.True(t, tt.checkResults(cards))
		})
	}
}

func ptr(s string) *string {
	return &s
}
