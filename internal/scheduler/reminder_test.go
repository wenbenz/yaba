package scheduler_test

import (
	"context"
	"testing"
	"time"
	"yaba/internal/database"
	"yaba/internal/model"
	"yaba/internal/scheduler"
	"yaba/internal/test/helper"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// mockMailer records sent emails.
type mockMailer struct {
	sent []string
}

func (m *mockMailer) Send(to, _, _ string) error {
	m.sent = append(m.sent, to)

	return nil
}

func TestReminderJobSendsOnce(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	ctx := context.Background()

	email := gofakeit.Email()
	user := &model.User{
		ID:                    uuid.New(),
		Username:              gofakeit.Username(),
		PasswordHash:          []byte("hash"),
		Email:                 &email,
		EmailRemindersEnabled: true,
	}

	require.NoError(t, database.CreateUser(ctx, pool, user))

	// Anniversary in 7 days
	cancelByDate := time.Now().UTC().AddDate(0, 0, 7)
	pm, err := createPaymentMethod(ctx, user.ID, cancelByDate)
	require.NoError(t, err)

	mailer := &mockMailer{}
	job := &scheduler.ReminderJob{Pool: pool, Mailer: mailer}

	require.NoError(t, job.Run(ctx))
	require.Len(t, mailer.sent, 1, "should send one reminder")
	require.Equal(t, email, mailer.sent[0])

	// Second run should not send again
	require.NoError(t, job.Run(ctx))
	require.Len(t, mailer.sent, 1, "should not resend")

	_ = pm
}

func TestReminderJobRespectsOptOut(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	ctx := context.Background()

	emailAddr := gofakeit.Email()
	user := &model.User{
		ID:                    uuid.New(),
		Username:              gofakeit.Username(),
		PasswordHash:          []byte("hash"),
		Email:                 &emailAddr,
		EmailRemindersEnabled: false,
	}

	require.NoError(t, database.CreateUser(ctx, pool, user))

	cancelByDate := time.Now().UTC().AddDate(0, 0, 7)
	_, err := createPaymentMethod(ctx, user.ID, cancelByDate)
	require.NoError(t, err)

	mailer := &mockMailer{}
	job := &scheduler.ReminderJob{Pool: pool, Mailer: mailer}

	require.NoError(t, job.Run(ctx))
	require.Empty(t, mailer.sent, "opted-out user should receive no email")
}

func TestReminderJobSkipsDistantRenewals(t *testing.T) {
	t.Parallel()

	pool := helper.GetTestPool()
	ctx := context.Background()

	emailAddr := gofakeit.Email()
	user := &model.User{
		ID:                    uuid.New(),
		Username:              gofakeit.Username(),
		PasswordHash:          []byte("hash"),
		Email:                 &emailAddr,
		EmailRemindersEnabled: true,
	}

	require.NoError(t, database.CreateUser(ctx, pool, user))

	// 60 days out — outside the 30-day window
	cancelByDate := time.Now().UTC().AddDate(0, 0, 60)
	_, err := createPaymentMethod(ctx, user.ID, cancelByDate)
	require.NoError(t, err)

	mailer := &mockMailer{}
	job := &scheduler.ReminderJob{Pool: pool, Mailer: mailer}

	require.NoError(t, job.Run(ctx))
	require.Empty(t, mailer.sent, "renewal too far out should not trigger reminder")
}

func createPaymentMethod(ctx context.Context, ownerID uuid.UUID, cancelByDate time.Time) (uuid.UUID, error) {
	id := uuid.New()
	cardType := uuid.New()

	_, err := helper.GetTestPool().Exec(ctx,
		`INSERT INTO payment_method (id, owner, display_name, card_type, cancel_by_date)
		 VALUES ($1, $2, $3, $4, $5)`,
		id, ownerID, gofakeit.CreditCardType(), cardType, cancelByDate,
	)

	return id, err
}
