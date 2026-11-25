package auth

import (
	"encoding/hex"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LogoutHandler struct {
	Pool *pgxpool.Pool
}

func NewLogoutHandler(pool *pgxpool.Pool) *LogoutHandler {
	return &LogoutHandler{
		Pool: pool,
	}
}

func (l *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ExpireSessionCookies(w, r)

	for _, sid := range r.CookiesNamed("sid") {
		decodedSID, err := hex.DecodeString(sid.Value)
		if err != nil || len(decodedSID) != sessionTokenSize {
			continue
		}

		if err = DeleteSessionToken(r.Context(), l.Pool, decodedSID); err != nil {
			log.Println("Error deleting session token:", err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

var _ http.Handler = (*LogoutHandler)(nil)
