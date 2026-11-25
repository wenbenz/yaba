package rpc_test

import (
	"database/sql"
	"testing"
	"time"

	"yaba/internal/model"
	"yaba/internal/rpc"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPaymentMethodToProto_NilModelIsNil(t *testing.T) {
	t.Parallel()

	require.Nil(t, rpc.PaymentMethodToProto(nil))
}

func TestPaymentMethodToProto_MapsAllFields(t *testing.T) {
	t.Parallel()

	acquired := time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)
	cancelBy := time.Date(2025, 3, 20, 0, 0, 0, 0, time.UTC)
	pm := &model.PaymentMethod{
		ID:           uuid.New(),
		Owner:        uuid.New(),
		DisplayName:  "Test Credit Card",
		AcquiredDate: sql.NullTime{Time: acquired, Valid: true},
		CancelByDate: sql.NullTime{Time: cancelBy, Valid: true},
		CardType:     uuid.New(),
	}

	got := rpc.PaymentMethodToProto(pm)
	require.NotNil(t, got)
	require.Equal(t, pm.ID.String(), got.GetId())
	require.Equal(t, "Test Credit Card", got.GetDisplayName())
	require.Equal(t, pm.CardType.String(), got.GetCardTypeId())
	require.NotNil(t, got.GetAcquiredDate())
	require.Equal(t, acquired, got.GetAcquiredDate().AsTime())
	require.NotNil(t, got.GetCancelByDate())
	require.Equal(t, cancelBy, got.GetCancelByDate().AsTime())
}

func TestPaymentMethodToProto_NilCardTypeAndNullDatesAreUnset(t *testing.T) {
	t.Parallel()

	pm := &model.PaymentMethod{
		ID:          uuid.New(),
		Owner:       uuid.New(),
		DisplayName: "Legacy Card",
		// Valid=false with a non-zero Time must still map to an unset timestamp.
		AcquiredDate: sql.NullTime{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Valid: false},
		CancelByDate: sql.NullTime{},
		CardType:     uuid.Nil,
	}

	got := rpc.PaymentMethodToProto(pm)
	require.NotNil(t, got)
	require.Equal(t, pm.ID.String(), got.GetId())
	require.Equal(t, "Legacy Card", got.GetDisplayName())
	require.Empty(t, got.GetCardTypeId())
	require.Nil(t, got.GetAcquiredDate())
	require.Nil(t, got.GetCancelByDate())
}
