package handlers

import (
	"log"
	"net/http"
	"yaba/internal/constants"
	"yaba/internal/platform"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UploadHandler struct {
	Pool *pgxpool.Pool
}

func (h UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(constants.CTXUser).(uuid.UUID)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	log.Println("handling upload for user " + user.String())

	file, _, err := r.FormFile("myFile")
	if err != nil {
		log.Println("Error Retrieving the File")
		log.Println(err)
		w.WriteHeader(http.StatusUnprocessableEntity)

		return
	}

	defer file.Close()

	err = platform.UploadSpendingsCSV(r.Context(), h.Pool, user, file)
	if err != nil {
		log.Println("Error reading CSV")
		log.Println(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(err.Error()))

		return
	}

	w.WriteHeader(http.StatusAccepted)
}
