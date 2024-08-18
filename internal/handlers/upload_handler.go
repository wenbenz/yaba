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

	if fileHeaders, ok := r.MultipartForm.File[expenditures]; ok {
		h.handleExpenditureCSV(r.Context(), fileHeaders, user, w)

		return
	}

	w.WriteHeader(http.StatusUnprocessableEntity)
}

func (h UploadHandler) handleExpenditureCSV(ctx context.Context,
	fileHeaders []*multipart.FileHeader, user uuid.UUID, w http.ResponseWriter) {
	wg := sync.WaitGroup{}
	failed := make(chan uploadError)

	// Upload expenditure for each file.
	for _, fh := range fileHeaders {
		go h.uploadExpenditure(ctx, fh, user, &wg, failed)
		wg.Add(1)
	}

	// Go routine to collect failures.
	failures := make(chan uploadErrors, 1)
	go func() {
		ues := uploadErrors{}
		for ue := range failed {
			ues[ue.filename] = ue.err
		}

		failures <- ues
		close(failures)
	}()

	// Once all the upload routines have completed, close the failures channel and check failures.
	wg.Wait()
	close(failed)

	allFailures := <-failures

	if len(allFailures) != 0 {
		w.WriteHeader(http.StatusBadRequest)

		if responseObject, err := json.Marshal(allFailures); err == nil {
			_, _ = w.Write(responseObject)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}

// Reads the CSV from fh and writes it to the expenditures table.
// Failures.
func (h UploadHandler) uploadExpenditure(ctx context.Context, fh *multipart.FileHeader, user uuid.UUID,
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

	if err = platform.UploadSpendingsCSV(ctx, h.Pool, user, file, filename); err != nil {
		log.Println("Error reading CSV: ", err)
		failed <- uploadError{filename: filename, err: err.Error()}

		return
	}
}
