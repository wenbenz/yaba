package handlers_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"
	"yaba/internal/constants"
	"yaba/internal/database"
	"yaba/internal/handlers"
	"yaba/internal/test/helper"
)

func TestUploadNoUser(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	pool := helper.GetTestPool()
	handler := handlers.UploadHandler{Pool: pool}

	request, err := UploadCSVRequest([]string{"testdata/spend.csv"})
	require.NoError(t, err)

	handler.ServeHTTP(w, request)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUploadCSV(t *testing.T) {
	t.Parallel()

	user := uuid.New()

	w := httptest.NewRecorder()

	pool := helper.GetTestPool()
	handler := handlers.UploadHandler{Pool: pool}

	request, err := UploadCSVRequest([]string{"testdata/spend.csv", "testdata/spend2.csv"})
	require.NoError(t, err)

	request = request.WithContext(context.WithValue(context.Background(), constants.CTXUser, user))

	handler.ServeHTTP(w, request)
	require.Equal(t, http.StatusOK, w.Code)

	// Check that rows in each csv exist
	date := time.Date(2006, time.July, 8, 0, 0, 0, 0, time.UTC)
	expenditures, err := database.ListExpenditures(context.Background(), pool, user, date, date, 100)
	require.NoError(t, err)
	require.Len(t, expenditures, 4)
}

func TestUploadNotCSV(t *testing.T) {
	t.Parallel()

	user := uuid.New()

	w := httptest.NewRecorder()

	pool := helper.GetTestPool()
	handler := handlers.UploadHandler{Pool: pool}

	request, err := UploadCSVRequest([]string{"testdata/file.txt"})
	require.NoError(t, err)

	request = request.WithContext(context.WithValue(context.Background(), constants.CTXUser, user))

	handler.ServeHTTP(w, request)
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"file.txt":"failed to import: `+
		`unrecognized column 'I'm a file.' in headers: `+
		`invalid input value: [I'm a file.]"}`,
		w.Body.String())
}

func TestUploadWrongKey(t *testing.T) {
	t.Parallel()

	user := uuid.New()

	w := httptest.NewRecorder()

	pool := helper.GetTestPool()
	handler := handlers.UploadHandler{Pool: pool}

	request, err := UploadFileRequest([]string{"testdata/file.txt"}, "text/csv", "foobar")
	require.NoError(t, err)

	request = request.WithContext(context.WithValue(context.Background(), constants.CTXUser, user))

	handler.ServeHTTP(w, request)
	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestUploadCSVPartialSuccess(t *testing.T) {
	t.Parallel()

	user := uuid.New()

	w := httptest.NewRecorder()

	pool := helper.GetTestPool()
	handler := handlers.UploadHandler{Pool: pool}

	request, err := UploadCSVRequest([]string{"testdata/spend.csv", "testdata/file.txt"})
	require.NoError(t, err)

	request = request.WithContext(context.WithValue(context.Background(), constants.CTXUser, user))

	handler.ServeHTTP(w, request)
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"file.txt":"failed to import: `+
		`unrecognized column 'I'm a file.' in headers: `+
		`invalid input value: [I'm a file.]"}`,
		w.Body.String())

	// Check that rows in the CSV
	date := time.Date(2006, time.July, 8, 0, 0, 0, 0, time.UTC)
	expenditures, err := database.ListExpenditures(context.Background(), pool, user, date, date, 100)
	require.NoError(t, err)
	require.Len(t, expenditures, 3)
}

func UploadCSVRequest(filepath []string) (*http.Request, error) {
	return UploadFileRequest(filepath, "text/csv", "expenditures[]")
}

func UploadFileRequest(filepaths []string, contentType, formKey string) (*http.Request, error) {
	var buf bytes.Buffer

	mw := multipart.NewWriter(&buf)
	defer mw.Close()

	var contentLength int64

	for _, filepath := range filepaths {
		csvFile, err := os.Open(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filepath, err)
		}

		stat, err := csvFile.Stat()
		if err != nil {
			return nil, fmt.Errorf("failed to stat file: %w", err)
		}

		hdr := make(textproto.MIMEHeader)
		cd := mime.FormatMediaType("form-data", map[string]string{
			"id":       formKey,
			"name":     formKey,
			"filename": filepath,
		})
		hdr.Set("Content-Disposition", cd)
		hdr.Set("Contnt-Type", contentType)
		hdr.Set("Content-Length", strconv.FormatInt(stat.Size(), 10))

		part, err := mw.CreatePart(hdr)
		if err != nil {
			return nil, fmt.Errorf("failed to create new form part: %w", err)
		}

		csvLen, err := io.Copy(part, csvFile)
		if err != nil {
			return nil, fmt.Errorf("failed to write form part: %w", err)
		}

		contentLength += csvLen
	}

	url, _ := url.Parse("http://localhost:8080/upload")

	header := make(http.Header)
	header.Set("Content-Type", mw.FormDataContentType())

	return &http.Request{
		Method:        http.MethodPost,
		URL:           url,
		Header:        header,
		Body:          io.NopCloser(&buf),
		ContentLength: contentLength,
	}, nil
}
