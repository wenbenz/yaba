package auth

import (
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"time"
	"yaba/internal/user"
)

type LoginHandler struct {
	Pool      *pgxpool.Pool
	LoginFunc func(context.Context, *pgxpool.Pool, string, string) (*uuid.UUID, error)
}

func CreateNewUserHandler(pool *pgxpool.Pool) *LoginHandler {
	return &LoginHandler{
		Pool:      pool,
		LoginFunc: user.CreateNewUser,
	}
}

func VerifyUserHandler(pool *pgxpool.Pool) *LoginHandler {
	return &LoginHandler{
		Pool:      pool,
		LoginFunc: user.VerifyUser,
	}
}

func (l *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	id, err := l.LoginFunc(r.Context(), l.Pool, username, password)
	if err != nil || id == nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// duplicate key constraint violation
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}

		return
	}

	// User is authenticated. Create a session
	token := NewSessionToken(*id, time.Hour)
	if err = SaveSessionToken(r.Context(), l.Pool, token); err != nil {
		log.Println("Error saving session token:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	cookie, err := BakeCookie(token, r.Host)
	if err != nil {
		log.Println("Error encoding session cookie:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

var _ http.Handler = (*LoginHandler)(nil)
