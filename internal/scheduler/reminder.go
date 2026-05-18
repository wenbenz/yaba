package scheduler

import (
	"context"
	"fmt"
	"time"
	"yaba/internal/database"
	"yaba/internal/email"

	"github.com/jackc/pgx/v5/pgxpool"
)

const reminderWindowDays = 30

// ReminderJob sends credit-card renewal reminders once per card per renewal year.
type ReminderJob struct {
	Pool   *pgxpool.Pool
	Mailer email.Mailer
}

func (j *ReminderJob) Run(ctx context.Context) error {
	targets, err := database.ListPaymentMethodsForReminder(ctx, j.Pool)
	if err != nil {
		return fmt.Errorf("list reminder targets: %w", err)
	}

	now := time.Now().UTC().Truncate(24 * time.Hour)

	for _, t := range targets {
		anniversary, renewalYear := nextAnniversary(t.CancelByDate, now)

		if anniversary.After(now.AddDate(0, 0, reminderWindowDays)) {
			continue
		}

		sent, err := database.WasNotificationSent(ctx, j.Pool, t.ID, renewalYear)
		if err != nil {
			return fmt.Errorf("check notification log for %s: %w", t.ID, err)
		}

		if sent {
			continue
		}

		if err = j.Mailer.Send(t.Email, reminderSubject(t.DisplayName), reminderBody(t.DisplayName, anniversary)); err != nil {
			return fmt.Errorf("send reminder to %s: %w", t.Email, err)
		}

		if err = database.LogNotificationSent(ctx, j.Pool, t.ID, t.Owner, renewalYear); err != nil {
			return fmt.Errorf("log notification for %s: %w", t.ID, err)
		}
	}

	return nil
}

// nextAnniversary returns the next upcoming anniversary of cancelByDate relative
// to now, and the calendar year of that anniversary (used as the dedup key).
func nextAnniversary(cancelByDate, now time.Time) (time.Time, int) {
	thisYear := time.Date(now.Year(), cancelByDate.Month(), cancelByDate.Day(), 0, 0, 0, 0, time.UTC)

	if !thisYear.Before(now) {
		return thisYear, now.Year()
	}

	return thisYear.AddDate(1, 0, 0), now.Year() + 1
}

func reminderSubject(displayName string) string {
	return fmt.Sprintf("Reminder: %s renewal coming up", displayName)
}

func reminderBody(displayName string, renewalDate time.Time) string {
	return fmt.Sprintf(
		`<p>Hi,</p>
<p>This is a reminder that your credit card <strong>%s</strong> is set to renew on <strong>%s</strong>.</p>
<p>If you'd like to cancel before the annual fee is charged, please do so before that date.</p>
<p>— YABA</p>`,
		displayName,
		renewalDate.Format("January 2, 2006"),
	)
}
