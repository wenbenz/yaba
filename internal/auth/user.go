package auth

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID `db:"id"`
	Username     string    `db:"username"`
	PasswordHash []byte    `db:"password_hash"`
}
