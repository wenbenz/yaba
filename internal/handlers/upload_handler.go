package handlers

import (
	"context"
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"sync"
	"yaba/internal/constants"
	"yaba/internal/platform"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMemory = 32 << 20
	expenditures  = "expenditures[]"
)

type UploadHandler struct {
	Pool *pgxpool.Pool
}

type uploadErrors map[string]string
type uploadError struct {
	filename string
	err      string
}

func (h UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(constants.CTXUser).(uuid.UUID)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	log.Println("handling upload for user " + user.String())

	if err := r.ParseMultipartForm(defaultMemory); err != nil { // 32 MB -- default from request.go
		log.Println("Error parsing the file: ", err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))

		return
	}

	failed := make(chan uploadError)

	wg := sync.WaitGroup{}

	if fileHeaders, ok := r.MultipartForm.File[expenditures]; ok {
		for _, fh := range fileHeaders {
			go uploadExpenditure(r.Context(), h.Pool, fh, user, &wg, failed)
			wg.Add(1)
		}

		// collect failures
		failures := make(chan uploadErrors, 1)
		go func() {
			ues := uploadErrors{}
			for ue := range failed {
				ues[ue.filename] = ue.err
			}

			failures <- ues
			close(failures)
		}()

		// close failures channel once all readers are done
		wg.Wait()
		close(failed)

		// build response from failures
		allFailures := <-failures
		if len(allFailures) != 0 {
			w.WriteHeader(http.StatusBadRequest)
			
			if responseObject, err := json.Marshal(allFailures); err == nil {
				_, _ = w.Write(responseObject)
			}

			return
		}

		w.WriteHeader(http.StatusOK)

		return
	}

	w.WriteHeader(http.StatusUnprocessableEntity)
}

// Reads the CSV from fh and writes it to the expenditures table.
// Failures.
func uploadExpenditure(ctx context.Context, pool *pgxpool.Pool, fh *multipart.FileHeader, user uuid.UUID,
	wg *sync.WaitGroup, failed chan<- uploadError) {
	defer wg.Done()

	filename := fh.Filename
	file, err := fh.Open()

	if err != nil {
		log.Println("Error opening the file: ", err)
		failed <- uploadError{filename: filename, err: err.Error()}

		return
	}

	defer file.Close()

	if err = platform.UploadSpendingsCSV(ctx, pool, user, file, filename); err != nil {
		log.Println("Error reading CSV: ", err)
		failed <- uploadError{filename: filename, err: err.Error()}

		return
	}
}
