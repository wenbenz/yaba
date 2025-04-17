package model

import (
	"github.com/google/uuid"
	"time"
	"yaba/internal/model"
)

// PaymentMethodToPaymentMethodResponse converts an internal payment method to a GraphQL response
func PaymentMethodToPaymentMethodResponse(pm *model.PaymentMethod) *PaymentMethod {
	if pm == nil {
		return nil
	}
	var acquired, cancel *string
	if !pm.AcquiredDate.IsZero() {
		date := pm.AcquiredDate.Format("2006-01-02")
		acquired = &date
	}
	if !pm.CancelByDate.IsZero() {
		date := pm.CancelByDate.Format("2006-01-02")
		cancel = &date
	}
	return &PaymentMethod{
		ID:           pm.ID.String(),
		DisplayName:  pm.DisplayName,
		AcquiredDate: acquired,
		CancelByDate: cancel,
		CardType:     pm.CardType.String(),
		Rewards:      RewardCardToRewardCardResponse(pm.Rewards),
	}
}

// PaymentMethodFromPaymentMethodInput converts a GraphQL input to an internal payment method
func PaymentMethodFromPaymentMethodInput(input PaymentMethodInput) (*model.PaymentMethod, error) {
	var acquired, cancel time.Time
	var err error
	if input.AcquiredDate != nil {
		if acquired, err = time.Parse("2006-01-02", *input.AcquiredDate); err != nil {
			return nil, err
		}
	}
	if input.CancelByDate != nil {
		if cancel, err = time.Parse("2006-01-02", *input.CancelByDate); err != nil {
			return nil, err
		}
	}
	cardType, err := uuid.Parse(*input.CardType)
	if err != nil {
		return nil, err
	}
	return &model.PaymentMethod{
		ID:           uuid.New(),
		DisplayName:  *input.DisplayName,
		AcquiredDate: acquired,
		CancelByDate: cancel,
		CardType:     cardType,
	}, nil
}
