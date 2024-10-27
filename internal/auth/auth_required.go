package auth

import (
	"github.com/google/uuid"
	"net/http"
	"yaba/internal/ctxutil"
)

type RequiresAuthenticatedUser struct {
	Intercepted http.Handler
}

func NewAuthRequired(h http.Handler) *RequiresAuthenticatedUser {
	return &RequiresAuthenticatedUser{
		Intercepted: h,
	}
}

func (a *RequiresAuthenticatedUser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ctxutil.GetUser(r.Context()) == uuid.Nil {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	a.Intercepted.ServeHTTP(w, r)
}

var _ http.Handler = (*RequiresAuthenticatedUser)(nil)
