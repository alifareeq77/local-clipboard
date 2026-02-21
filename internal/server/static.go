package server

import (
	"net/http"
	"os"
	"path"
	"strings"
)

// spaHandler serves static files from rootDir and falls back to index.html for SPA routing.
// If rootDir is empty or the directory does not exist, it serves the embedded indexHTML.
type spaHandler struct {
	rootDir string
	embed  []byte
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	p := r.URL.Path
	if p == "" || p == "/" {
		p = "/index.html"
	}
	if strings.Contains(p, "..") {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if h.rootDir != "" {
		fpath := path.Join(h.rootDir, strings.TrimPrefix(path.Clean(p), "/"))
		if f, err := os.Open(fpath); err == nil {
			st, _ := f.Stat()
			if st != nil && !st.IsDir() {
				f.Close()
				http.ServeFile(w, r, fpath)
				return
			}
			f.Close()
		}
		indexPath := path.Join(h.rootDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(h.embed)
}
