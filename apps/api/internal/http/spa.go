package httpserver

import (
	"net/http"
	"os"
	"path/filepath"
)

// spaHandler serves static files out of webRoot, falling back to index.html
// for any path that doesn't match a real file - standard SPA client-side
// routing support.
func spaHandler(webRoot string) http.HandlerFunc {
	fs := http.FileServer(http.Dir(webRoot))
	return func(w http.ResponseWriter, r *http.Request) {
		full := filepath.Join(webRoot, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(webRoot, "index.html"))
	}
}
