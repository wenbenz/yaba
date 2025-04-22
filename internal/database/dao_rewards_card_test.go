package database_test

import (
	"fmt"
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
				ID:         uuid.New(),
				Name:       "Cash Back",
				Region:     "Canada",
				Version:    1,
				Issuer:     "Chase",
				RewardType: "cash",
			},
		},
		{
			name: "valid reward card with different values",
			reward: &model.RewardCard{
				ID:      uuid.New(),
				Name:    "Travel Points",
				Region:  "Canada",
				Version: 1,
				Issuer:  "American Express",

				RewardType: "points",
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
				ID:      uuid.New(),
				Region:  "USA",
				Version: 1,
				Issuer:  "Chase",

				RewardType: "cash",
			},
			expectError: true,
		},
		{
			name: "invalid reward card - missing region",
			reward: &model.RewardCard{
				ID:      uuid.New(),
				Name:    "Cash Back",
				Version: 1,
				Issuer:  "Chase",

				RewardType: "cash",
			},
			expectError: true,
		},
		{
			name: "invalid reward card - missing issuer",
			reward: &model.RewardCard{
				ID:      uuid.New(),
				Name:    "Cash Back",
				Region:  "Canada",
				Version: 1,

				RewardType: "cash",
			},
			expectError: true,
		},
		{
			name: "invalid reward card - missing reward type",
			reward: &model.RewardCard{
				ID:      uuid.New(),
				Name:    "Cash Back",
				Region:  "Canada",
				Version: 1,
				Issuer:  "Chase",
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
		})
	}
}

func TestCreateRewardCardVersioning(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	ctx := t.Context()

	card1 := &model.RewardCard{
		ID:      uuid.New(),
		Name:    "Freedom Flex",
		Version: 1,
		Issuer:  "Chase",
		Region:  "USA",

		RewardType: "cash",
	}

	card2 := &model.RewardCard{
		ID:      uuid.New(),
		Name:    "Freedom Flex",
		Version: 2,
		Issuer:  "Chase",
		Region:  "USA",

		RewardType: "cash",
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
			ID:      uuid.New(),
			Name:    cashBackPlus,
			Region:  canada,
			Version: 1,
			Issuer:  chase,

			RewardType: "cash",
		},
		{
			ID:      uuid.New(),
			Name:    travelPoints,
			Region:  usa,
			Version: 1,
			Issuer:  amex,

			RewardType: "points",
		},
		{
			ID:      uuid.New(),
			Name:    "Cash Back Basic",
			Region:  canada,
			Version: 1,
			Issuer:  td,

			RewardType: "cash",
		},
		{
			ID:      uuid.New(),
			Name:    "Premium Card",
			Region:  usa,
			Version: 1,
			Issuer:  chase,

			RewardType: "cash",
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

			cards, err := database.ListRewardCards(ctx, pool, tt.issuer, tt.cardName, tt.region, nil, nil)
			require.NoError(t, err)
			require.Len(t, cards, tt.expectedLen)
			require.True(t, tt.checkResults(cards))
		})
	}
}

//nolint:paralleltest
func TestListRewardCardsPagination(t *testing.T) {
	pool := helper.GetTestPool()
	ctx := t.Context()

	// Clear the table before testing
	_, _ = pool.Exec(ctx, "TRUNCATE TABLE rewards_card")

	// Create 15 test cards
	cards := make([]*model.RewardCard, 15)
	for i := range 15 {
		cards[i] = &model.RewardCard{
			ID:      uuid.New(),
			Name:    fmt.Sprintf("Card %02d", i+1),
			Region:  "Canada",
			Version: 1,
			Issuer:  "Test Bank",

			RewardType: "cash",
		}
		err := database.CreateRewardCard(ctx, pool, cards[i])
		require.NoError(t, err)
	}

	tests := []struct {
		name         string
		limit        *int
		offset       *int
		expectedLen  int
		checkResults func([]*model.RewardCard) bool
	}{
		{
			name:        "default limit (10)",
			expectedLen: 10,
			checkResults: func(results []*model.RewardCard) bool {
				return len(results) == 10 && results[0].Name == "Card 01"
			},
		},
		{
			name:        "custom limit",
			limit:       ptr(5),
			expectedLen: 5,
			checkResults: func(results []*model.RewardCard) bool {
				return len(results) == 5 && results[0].Name == "Card 01"
			},
		},
		{
			name:        "with offset",
			limit:       ptr(5),
			offset:      ptr(5),
			expectedLen: 5,
			checkResults: func(results []*model.RewardCard) bool {
				return len(results) == 5 && results[0].Name == "Card 06"
			},
		},
		{
			name:        "offset beyond data",
			limit:       ptr(5),
			offset:      ptr(20),
			expectedLen: 0,
			checkResults: func(results []*model.RewardCard) bool {
				return len(results) == 0
			},
		},
		{
			name:        "large limit",
			limit:       ptr(20),
			expectedLen: 15,
			checkResults: func(results []*model.RewardCard) bool {
				return len(results) == 15
			},
		},
	}

	for _, tt := range tests {
		results, err := database.ListRewardCards(ctx, pool, nil, nil, nil, tt.limit, tt.offset)
		require.NoError(t, err)
		require.Len(t, results, tt.expectedLen)
		require.True(t, tt.checkResults(results))
	}
}

func ptr[T interface{}](s T) *T {
	return &s
}
