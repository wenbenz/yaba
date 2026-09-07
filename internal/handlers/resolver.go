package handlers

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"yaba/internal/email"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Pool   *pgxpool.Pool
	Mailer email.Mailer
}
