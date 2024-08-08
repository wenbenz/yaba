package handlers_test

import (
	"bytes"
	"context"
	"fmt"
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
	"yaba/internal/constants"
	"yaba/internal/handlers"
	"yaba/test/helper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUploadNoUser(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	pool := helper.GetTestPool()
	handler := handlers.UploadHandler{Pool: pool}

	request, err := UploadCSVRequest("testdata/spend.csv")
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

	request, err := UploadCSVRequest("testdata/spend.csv")
	require.NoError(t, err)

	request = request.WithContext(context.WithValue(context.Background(), constants.CTXUser, user))

	handler.ServeHTTP(w, request)
	require.Equal(t, http.StatusAccepted, w.Code)
}

func TestUploadNotCSV(t *testing.T) {
	t.Parallel()

	user := uuid.New()

	w := httptest.NewRecorder()

	pool := helper.GetTestPool()
	handler := handlers.UploadHandler{Pool: pool}

	request, err := UploadCSVRequest("testdata/file.txt")
	require.NoError(t, err)

	request = request.WithContext(context.WithValue(context.Background(), constants.CTXUser, user))

	handler.ServeHTTP(w, request)
	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestUploadWrongKey(t *testing.T) {
	t.Parallel()

	user := uuid.New()

	w := httptest.NewRecorder()

	pool := helper.GetTestPool()
	handler := handlers.UploadHandler{Pool: pool}

	request, err := UploadFileRequest("testdata/file.txt", "text/csv", "foobar")
	require.NoError(t, err)

	request = request.WithContext(context.WithValue(context.Background(), constants.CTXUser, user))

	handler.ServeHTTP(w, request)
	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func UploadCSVRequest(filepath string) (*http.Request, error) {
	return UploadFileRequest(filepath, "text/csv", "myFile")
}

func UploadFileRequest(filepath, contentType, formKey string) (*http.Request, error) {
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

	var buf bytes.Buffer

	mw := multipart.NewWriter(&buf)
	defer mw.Close()

	part, err := mw.CreatePart(hdr)
	if err != nil {
		return nil, fmt.Errorf("failed to create new form part: %w", err)
	}

	csvLen, err := io.Copy(part, csvFile)
	if err != nil {
		return nil, fmt.Errorf("failed to write form part: %w", err)
	}

	url, _ := url.Parse("http://localhost:8080/upload")

	header := make(http.Header)
	header.Set("Content-Type", mw.FormDataContentType())

	return &http.Request{
		Method:        http.MethodPost,
		URL:           url,
		Header:        header,
		Body:          io.NopCloser(&buf),
		ContentLength: csvLen,
	}, nil
}
