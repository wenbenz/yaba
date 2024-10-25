package auth

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

type LoginHandler struct {
	Pool       *pgxpool.Pool
	LoginFunc  func(context.Context, *pgxpool.Pool, string, string) (*uuid.UUID, error)
	FailStatus int
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
		w.WriteHeader(l.FailStatus)

		return
	}

	// User is authenticated. Create a session
	token := NewSessionToken(*id, time.Hour)
	if err := SaveSessionToken(r.Context(), l.Pool, token); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	http.SetCookie(w, bakeCookie(token))
	http.Redirect(w, r, "/", http.StatusFound)
}

var _ http.Handler = (*LoginHandler)(nil)
