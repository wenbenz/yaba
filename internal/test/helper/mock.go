package helper

import (
	"fmt"
	"slices"
	"time"
	"yaba/internal/database"
	"yaba/internal/model"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

type TestDataGenerator struct {
	Owner uuid.UUID

	Expenditures      []*model.Expenditure
	RewardsCards      []*model.RewardCard
	PaymentMethods    []*model.PaymentMethod
	PaymentMethodsMap map[uuid.UUID]*model.PaymentMethod

	faker *gofakeit.Faker
}

func NewTestDataGenerator(owner uuid.UUID, seed uint64) *TestDataGenerator {
	return &TestDataGenerator{
		Owner:             owner,
		faker:             gofakeit.New(seed),
		PaymentMethodsMap: make(map[uuid.UUID]*model.PaymentMethod),
	}
}

func (g *TestDataGenerator) PersistAll(ctx context.Context, pool *pgxpool.Pool) error {
	for _, card := range g.RewardsCards {
		if err := database.CreateRewardCard(ctx, pool, card); err != nil {
			return fmt.Errorf("failed to create reward card: %w", err)
		}
	}

	for _, method := range g.PaymentMethods {
		if err := database.CreatePaymentMethod(ctx, pool, method); err != nil {
			return fmt.Errorf("failed to create payment method: %w", err)
		}

		g.PaymentMethodsMap[method.ID] = method
	}

	if err := database.PersistExpenditures(ctx, pool, g.Expenditures); err != nil {
		return fmt.Errorf("failed to create expenditures: %w", err)
	}

	slices.SortFunc(g.Expenditures, func(a, b *model.Expenditure) int {
		dateComp := -a.Date.Compare(b.Date)
		if dateComp == 0 {
			return b.ID - a.ID
		}

		return dateComp
	})

	return nil
}

func (g *TestDataGenerator) GenerateRewardsCards(n int) []*model.RewardCard {
	g.RewardsCards = make([]*model.RewardCard, n)
	for i := range n {
		g.RewardsCards[i] = &model.RewardCard{
			ID:         uuid.New(),
			Name:       g.faker.NounAbstract() + " " + g.faker.CreditCardType(),
			Region:     g.faker.Country(),
			Issuer:     g.faker.Company(),
			RewardType: g.rewardType(),
		}
	}

	return g.RewardsCards
}

func (g *TestDataGenerator) GeneratePaymentMethods(n int, owner uuid.UUID) []*model.PaymentMethod {
	if g.RewardsCards == nil {
		g.GenerateRewardsCards(n/2 + 1)
	}

	g.PaymentMethods = make([]*model.PaymentMethod, n)
	for i := range n {
		g.PaymentMethods[i] = &model.PaymentMethod{
			ID:          uuid.New(),
			Owner:       owner,
			DisplayName: g.faker.CreditCardType() + fmt.Sprintf(" %d", i),
		}
	}

	return g.PaymentMethods
}

func (g *TestDataGenerator) GenerateExpenditures(
	n int,
	owner uuid.UUID,
	startDate, endDate time.Time,
) []*model.Expenditure {
	if g.PaymentMethods == nil {
		g.GeneratePaymentMethods(3, owner)
	}

	g.Expenditures = make([]*model.Expenditure, n)
	for i := range g.Expenditures {
		g.Expenditures[i] = &model.Expenditure{
			ID:     i + 1,
			Owner:  owner,
			Name:   g.faker.BeerStyle(),
			Amount: g.faker.Float64Range(0.01, 100),
			Date: g.faker.DateRange(startDate, endDate.AddDate(0, 0, 1)).
				UTC().
				Truncate(24 * time.Hour),
			Method:         g.paymentMethod().ID,
			BudgetCategory: g.budgetCategory(),
			RewardCategory: g.rewardCategory(),
			Comment:        g.faker.HipsterSentence(3), //nolint:mnd
			CreatedTime:    g.faker.DateRange(startDate.UTC(), endDate.AddDate(0, 0, 1)),
			Source:         g.faker.CarMaker() + ".csv",
		}
	}

	for i, e := range g.Expenditures {
		e.ID = i + 1
	}

	return g.Expenditures
}

func (g *TestDataGenerator) paymentMethod() *model.PaymentMethod {
	return g.PaymentMethods[g.faker.IntN(len(g.PaymentMethods))]
}

func (g *TestDataGenerator) rewardType() string {
	return g.faker.RandomString([]string{
		"Cashback",
		"Amex Membership Rewards",
		"Aeroplan",
		"Air Miles",
		"Scene+",
		"PC Optimum",
		"CIBC Aventura",
		"RBC Avion",
		"TD Rewards",
		"WestJet Dollars",
		"Triangle Rewards",
	})
}

func (g *TestDataGenerator) rewardCategory() string {
	return g.faker.RandomString([]string{
		"DRUG_STORE",
		"ENTERTAINMENT",
		"FURNITURE",
		"GAS",
		"GROCERY",
		"HOME_IMPROVEMENT",
		"HOTEL",
		"PUBLIC_TRANSPORTATION",
		"RECURRING_BILL",
		"RESTAURANT",
	})
}

func (g *TestDataGenerator) budgetCategory() string {
	return g.faker.RandomString([]string{
		"Rent",
		"Groceries",
		"Utilities",
		"Transportation",
		"Entertainment",
		"Dining Out",
		"Clothing",
		"Health & Fitness",
		"Travel",
		"Education",
		"Miscellaneous",
	})
}
