package auth

import (
	"encoding/hex"
	"net/http"
	"yaba/internal/ctxutil"

	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionInterceptor struct {
	Pool        *pgxpool.Pool
	Intercepted http.Handler
}

var _ http.Handler = &SessionInterceptor{}

func (si *SessionInterceptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Browsers may send more than one sid (e.g. a stale legacy Domain cookie); use the first valid one.
	for _, sid := range r.CookiesNamed("sid") {
		if token := si.lookup(r.Context(), sid.Value); token != nil {
			ctx := context.WithValue(r.Context(), ctxutil.CTXSID, token.ID)
			ctx = context.WithValue(ctx, ctxutil.CTXUser, token.User)
			r = r.WithContext(ctx)

			break
		}
	}

	si.Intercepted.ServeHTTP(w, r)
}

func (si *SessionInterceptor) lookup(ctx context.Context, sid string) *Token {
	decodedSID, err := hex.DecodeString(sid)
	if err != nil || len(decodedSID) != sessionTokenSize {
		return nil
	}

	token, err := GetSessionToken(ctx, si.Pool, decodedSID)
	if err != nil {
		return nil
	}

	return token
}
