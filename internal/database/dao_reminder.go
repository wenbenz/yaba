package database

import (
	"context"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ReminderTarget is a payment method whose owner has opted into email reminders.
type ReminderTarget struct {
	ID           uuid.UUID `db:"id"`
	Owner        uuid.UUID `db:"owner"`
	DisplayName  string    `db:"display_name"`
	CancelByDate time.Time `db:"cancel_by_date"`
	Email        string    `db:"email"`
}

// ListPaymentMethodsForReminder returns all payment methods whose owner has
// an email address and has enabled email reminders, and whose cancel_by_date
// is set. Filtering by date window is done in the caller.
func ListPaymentMethodsForReminder(ctx context.Context, pool *pgxpool.Pool) ([]*ReminderTarget, error) {
	query := `
		SELECT pm.id, pm.owner, pm.display_name, pm.cancel_by_date, up.email
		FROM payment_method pm
		JOIN user_profile up ON pm.owner = up.id
		WHERE up.email IS NOT NULL
		  AND up.email_reminders_enabled = true
		  AND pm.cancel_by_date IS NOT NULL`

	var targets []*ReminderTarget
	if err := pgxscan.Select(ctx, pool, &targets, query); err != nil {
		return nil, fmt.Errorf("failed to list reminder targets: %w", err)
	}

	return targets, nil
}
