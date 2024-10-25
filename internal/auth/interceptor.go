package auth

import (
	"encoding/hex"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
	"net/http"
	"yaba/internal/ctxutil"
)

type Interceptor struct {
	Pool        *pgxpool.Pool
	Intercepted http.Handler
}

var _ http.Handler = &Interceptor{}

func (l *Interceptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	SID, _ := r.Cookie("sid")
	if SID == nil {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	decodedSID, err := hex.DecodeString(SID.Value)
	if err != nil || len(decodedSID) != 16 {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	token, err := GetSessionToken(ctx, l.Pool, uuid.UUID(decodedSID))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	ctx = context.WithValue(ctx, ctxutil.CTXSID, token.ID)
	ctx = context.WithValue(ctx, ctxutil.CTXUser, token.User)
	r = r.WithContext(ctx)

	l.Intercepted.ServeHTTP(w, r)
}
