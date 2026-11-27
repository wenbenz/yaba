package model

import (
	"database/sql"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID      `db:"id"`
	Username     string         `db:"username"`
	PasswordHash []byte         `db:"password_hash"`
	Email        sql.NullString `db:"email"`
}
