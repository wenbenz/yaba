package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"os"
	"time"
)

type Token struct {
	ID      []byte    `db:"id"`
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
		ID:      value,
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

func GetSessionToken(ctx context.Context, pool *pgxpool.Pool, id []byte) (*Token, error) {
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

	return token, nil
}

func DeleteSessionToken(ctx context.Context, pool *pgxpool.Pool, id []byte) error {
	sql, args, err := squirrel.Delete("token").Where(squirrel.Eq{"id": id}).PlaceholderFormat(squirrel.Dollar).ToSql()
	if err == nil {
		if _, err = pool.Exec(ctx, sql, args...); err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to delete session token: %w", err)
}

func BakeCookie(token *Token, domain string) (*http.Cookie, error) {
	secure := os.Getenv("INSECURE_COOKIE") != "true"

	return &http.Cookie{
		Name:     "sid",
		Value:    hex.EncodeToString(token.ID),
		Path:     "/",
		Domain:   domain,
		Expires:  token.Expires,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}, nil
}
