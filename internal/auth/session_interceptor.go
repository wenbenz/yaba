package auth

import (
	"encoding/hex"
	"net/http"
	"yaba/internal/ctxutil"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

type SessionInterceptor struct {
	Pool        *pgxpool.Pool
	Intercepted http.Handler
}

var _ http.Handler = &SessionInterceptor{}

func (si *SessionInterceptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	SID, _ := r.Cookie("sid")
	if SID != nil {
		r = r.WithContext(si.setContext(r.Context(), SID.Value))
	}

	si.Intercepted.ServeHTTP(w, r)
}

func (si *SessionInterceptor) setContext(ctx context.Context, sid string) context.Context {
	decodedSID, err := hex.DecodeString(sid)
	if err != nil || len(decodedSID) != sessionTokenSize {
		return ctx
	}

	token, err := GetSessionToken(ctx, si.Pool, decodedSID)
	if err != nil {
		return ctx
	}

	ctx = context.WithValue(ctx, ctxutil.CTXSID, token.ID)
	ctx = context.WithValue(ctx, ctxutil.CTXUser, token.User)

	return ctx
}
