package auth

import (
	"encoding/hex"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
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
	SID, err := r.Cookie("sid")

	if err != nil || SID == nil {
		http.RedirectHandler("/", http.StatusTemporaryRedirect)

		return
	}

	decodedSID, err := hex.DecodeString(SID.Value)
	if err != nil || len(decodedSID) != sessionTokenSize {
		http.RedirectHandler("/", http.StatusTemporaryRedirect)

		return
	}

	if err = DeleteSessionToken(r.Context(), l.Pool, decodedSID); err != nil {
		log.Println("Error deleting session token:", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	// Clear the session cookie
	cookie := &http.Cookie{
		Name:   "sid",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	http.RedirectHandler("/", http.StatusTemporaryRedirect)
}

var _ http.Handler = (*LogoutHandler)(nil)
