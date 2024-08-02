package budget_test

import (
	"testing"
	"yaba/internal/budget"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBudgetTransitivity(t *testing.T) {
	t.Parallel()

	owner, err := uuid.NewRandom()
	require.NoError(t, err)

	b := budget.NewZeroBasedBudget(owner, "testBudget")
	b.SetBudgetIncome("income", 5000.)
	b.SetFixedExpense("rent", 2000.)
	b.SetFixedExpense("groceries", 1000.)
	b.SetSlackExpense("savings")

	require.EqualValues(t, b, b.ToExternal().ToInternal())
}
