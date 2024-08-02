package platform_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"
	"yaba/internal/budget"
	"yaba/internal/database"
	"yaba/internal/platform"
	"yaba/test/helper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBasicBudgetOperations(t *testing.T) {
	t.Parallel()

	pool, closeContainer := helper.SetupTestContainerAndInitPool()
	defer closeContainer()

	path := "testdata/budget.json"
	f, err := os.Open(path)
	require.NoError(t, err)

	owner, err := uuid.Parse("b52d2560-9e3b-407c-8593-beac1ea851ff")
	require.NoError(t, err)

	user, err := uuid.NewRandom()
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), "user", user.String())

	// Upload json.
	require.NoError(t, platform.UploadBudget(ctx, pool, f))

	// Check that the budget has been correctly saved.
	fetched, err := database.GetBudgets(ctx, pool, owner)
	require.NoError(t, err)
	require.Len(t, fetched, 1)

	b := fetched[0]
	require.NotNil(t, b.ID)
	require.Equal(t, owner, b.Owner)
	require.Equal(t, "Name", b.Name)
	require.Equal(t, budget.ZeroBased, b.Strategy)
	require.Len(t, b.Incomes, 1)
	require.InDelta(t, 6000., b.Incomes["work"].Amount, .00001)
	require.Len(t, b.Expenses, 4)
	require.InDelta(t, 2000., b.Expenses["rent"].Amount, .00001)
	require.True(t, b.Expenses["rent"].Fixed)
	require.False(t, b.Expenses["rent"].Slack)
	require.InDelta(t, 1000., b.Expenses["food"].Amount, .00001)
	require.True(t, b.Expenses["food"].Fixed)
	require.False(t, b.Expenses["food"].Slack)
	require.InDelta(t, 300., b.Expenses["hobbies"].Amount, .00001)
	require.True(t, b.Expenses["hobbies"].Fixed)
	require.False(t, b.Expenses["hobbies"].Slack)
	require.InDelta(t, 0., b.Expenses["savings"].Amount, .00001)
	require.False(t, b.Expenses["savings"].Fixed)
	require.True(t, b.Expenses["savings"].Slack)

	// modify the budget and re-upload
	b.RemoveExpense("hobbies")
	jsonData, err := json.Marshal(b.ToExternal())
	require.NoError(t, err)
	buffer := bytes.NewBuffer(jsonData)
	require.NoError(t, platform.UploadBudget(ctx, pool, buffer))

	// since ID is included, this should be an update
	fetched, err = database.GetBudgets(ctx, pool, owner)
	require.NoError(t, err)
	require.Len(t, fetched, 1)
	require.EqualValues(t, b, fetched[0])
}
