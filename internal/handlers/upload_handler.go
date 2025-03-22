package handlers

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"mime/multipart"
	"net/http"
	"sync"
	"yaba/internal/ctxutil"
	"yaba/internal/import"
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
	err      error
}

func (h UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(ctxutil.CTXUser).(uuid.UUID)
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
	failed := make(chan *uploadError)

	// Upload expenditure for each file.
	for _, fh := range fileHeaders {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := h.uploadExpenditure(ctx, fh, user); err != nil {
				failed <- err
			}
		}()
	}

	// Go routine to collect failures.
	failures := make(chan uploadErrors, 1)
	go func() {
		ues := uploadErrors{}
		for ue := range failed {
			ues[ue.filename] = ue.err.Error()
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
func (h UploadHandler) uploadExpenditure(ctx context.Context, fh *multipart.FileHeader, user uuid.UUID) *uploadError {
	filename := fh.Filename
	file, err := fh.Open()

	if err != nil {
		return &uploadError{filename: filename, err: err}
	}

	defer file.Close()

	if err = importer.UploadSpendingsCSV(ctx, h.Pool, user, file, filename); err != nil {
		return &uploadError{filename: filename, err: err}
	}

	return nil
}
