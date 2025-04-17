package model

import (
	"github.com/google/uuid"
	"time"
)

type PaymentMethod struct {
	ID           uuid.UUID `db:"id"`
	Owner        uuid.UUID `db:"owner"`
	DisplayName  string    `db:"display_name"`
	AcquiredDate time.Time `db:"acquired_date"`
	CancelByDate time.Time `db:"cancel_by_date"`
	CardType     uuid.UUID `db:"card_type"`
	Rewards      *RewardCard
}

type RewardCard struct {
	ID              uuid.UUID `db:"id"`
	Name            string    `db:"name"`
	Region          string    `db:"region"`
	Issuer          string    `db:"issuer"`
	Version         int       `db:"version"`
	RewardRate      float64   `db:"reward_rate"`
	RewardType      string    `db:"reward_type"`
	RewardCashValue float64   `db:"reward_cash_value"`
}
