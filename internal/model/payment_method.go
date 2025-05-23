package model

import (
	"database/sql"
	"github.com/google/uuid"
)

type PaymentMethod struct {
	ID           uuid.UUID    `db:"id"`
	Owner        uuid.UUID    `db:"owner"`
	DisplayName  string       `db:"display_name"`
	AcquiredDate sql.NullTime `db:"acquired_date"`
	CancelByDate sql.NullTime `db:"cancel_by_date"`
	CardType     uuid.UUID    `db:"card_type"`
	Rewards      *RewardCard
}

type RewardCard struct {
	ID               uuid.UUID `db:"id"`
	Name             string    `db:"name"`
	Region           string    `db:"region"`
	Issuer           string    `db:"issuer"`
	Version          int       `db:"version"`
	RewardType       string    `db:"reward_type"`
	RewardCategories []*RewardCategory
}

type RewardCategory struct {
	CardID   uuid.UUID `db:"card_id"`
	Category string    `db:"category"`
	Rate     float64   `db:"reward_rate"`
}
