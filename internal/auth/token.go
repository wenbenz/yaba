package auth

import (
	"crypto/rand"
	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
	"log"
	"time"
)

type Token struct {
	ID      uuid.UUID `db:"id"`
	User    uuid.UUID `db:"user_id"`
	Type    string    `db:"type"`
	Created time.Time `db:"created"`
	Expires time.Time `db:"expires"`
}

func NewSessionToken(user uuid.UUID, ttl time.Duration) *Token {
	value := make([]byte, 32)
	_, err := rand.Read(value)
	if err != nil {
		log.Println("Error generating random value:", err)
		return nil
	}

	return &Token{
		ID:      uuid.New(),
		User:    user,
		Type:    "SESSION",
		Created: time.Now(),
		Expires: time.Now().Add(ttl),
	}
}

func SaveSessionToken(ctx context.Context, pool *pgxpool.Pool, token *Token) error {
	sql, args, err := squirrel.Insert("token").
		Values(token.ID, token.User, token.Type, token.Created, token.Expires).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err == nil {
		_, err = pool.Exec(ctx, sql, args...)
	}

	return err
}

func GetSessionToken(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*Token, error) {
	sql, args, err := squirrel.Select("*").From("token").Where(squirrel.Eq{"id": id}).PlaceholderFormat(squirrel.Dollar).ToSql()
	token := &Token{}
	if err == nil {
		err = pgxscan.Get(ctx, pool, token, sql, args...)
	}

	if token.ID == uuid.Nil {
		token = nil
	}

	return token, err
}

func DeleteSessionToken(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) error {
	sql, args, err := squirrel.Delete("token").Where(squirrel.Eq{"id": id}).PlaceholderFormat(squirrel.Dollar).ToSql()
	if err == nil {
		_, err = pool.Exec(ctx, sql, args...)
	}

	return err
}
