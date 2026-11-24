package ui_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"yaba/ui"

	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	assets := fstest.MapFS{
		"index.html":    {Data: []byte("index")},
		"assets/app.js": {Data: []byte("app")},
	}
	handler := ui.Handler(assets)

	for path, want := range map[string]string{
		"/":              "index",
		"/assets/app.js": "app",
		"/profile":       "index",
		"/budget/123":    "index",
	} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))

		require.Equal(t, http.StatusOK, rec.Code, path)
		require.Equal(t, want, rec.Body.String(), path)
	}
}

func TestAssets(t *testing.T) {
	t.Parallel()

	_, err := ui.Assets()
	require.NoError(t, err)
}
