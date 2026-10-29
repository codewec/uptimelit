package server

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// serveFrontend serves Vite assets and falls back to the SPA for page routes.
// API routes are handled separately by chi and never return the application shell.
func (s *Server) serveFrontend(w http.ResponseWriter, r *http.Request) {
	if s.frontend == nil {
		http.Error(w, "Frontend is not built. Run make build or start the Vite development server.", http.StatusServiceUnavailable)
		return
	}
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if strings.HasPrefix(name, ".") {
		http.NotFound(w, r)
		return
	}
	if name != "" && name != "." && name != "index.html" {
		info, err := fs.Stat(s.frontend, name)
		if err == nil && !info.IsDir() {
			if strings.HasPrefix(name, "assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			http.FileServer(http.FS(s.frontend)).ServeHTTP(w, r)
			return
		}
		if path.Ext(name) != "" || strings.HasPrefix(name, "assets/") {
			http.NotFound(w, r)
			return
		}
	}
	index, err := fs.ReadFile(s.frontend, "index.html")
	if err != nil {
		http.Error(w, "Frontend is not built. Run make build or start the Vite development server.", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(index)
}
