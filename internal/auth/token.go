package auth

import (
	"crypto/rand"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"time"
)

type Token struct {
	ID      uuid.UUID `db:"id"`
	User    uuid.UUID `db:"user_id"`
	Type    string    `db:"type"`
	Created time.Time `db:"created"`
	Expires time.Time `db:"expires"`
}

const sessionTokenSize = 32

func NewSessionToken(user uuid.UUID, ttl time.Duration) *Token {
	value := make([]byte, sessionTokenSize)
	if _, err := rand.Read(value); err != nil {
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

	if err != nil {
		return fmt.Errorf("failed to persist session token: %w", err)
	}

	return nil
}

func GetSessionToken(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*Token, error) {
	sql, args, err := squirrel.Select("*").
		From("token").
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	token := &Token{}
	if err == nil {
		err = pgxscan.Get(ctx, pool, token, sql, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve session token: %w", err)
	}

	if token.ID == uuid.Nil {
		return nil, fmt.Errorf("invalid session token")
	}

	return token, nil
}

func DeleteSessionToken(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) error {
	sql, args, err := squirrel.Delete("token").Where(squirrel.Eq{"id": id}).PlaceholderFormat(squirrel.Dollar).ToSql()
	if err == nil {
		if _, err = pool.Exec(ctx, sql, args...); err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to delete session token: %w", err)
}

func bakeCookie(token *Token) *http.Cookie {
	return &http.Cookie{
		Name:     "sid",
		Value:    token.ID.String(),
		Path:     "/",
		Domain:   "localhost",
		Expires:  token.Expires,
		Secure:   true,
		HttpOnly: true,
	}
}
