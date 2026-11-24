// Package ui embeds the built React app and serves it as a single-page app.
package ui

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// dist holds the output of `yarn build`. Only .gitkeep is present until the UI is built.
//
//go:embed all:dist
var dist embed.FS

// Assets returns the embedded UI build rooted at dist/.
func Assets() (fs.FS, error) {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		return nil, fmt.Errorf("could not open embedded ui: %w", err)
	}

	return sub, nil
}

// Handler serves files from assets, falling back to index.html so client-side routes can be deep-linked.
func Handler(assets fs.FS) http.Handler {
	fileServer := http.FileServerFS(assets)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "."
		}

		if _, err := fs.Stat(assets, name); err != nil {
			http.ServeFileFS(w, r, assets, "index.html")

			return
		}

		fileServer.ServeHTTP(w, r)
	})
}
