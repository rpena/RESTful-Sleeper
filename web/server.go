package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed index.html dashboard.js sleeper-dashboard-core.js sleeper-dashboard-card.js
var assets embed.FS

func NewHandler(api http.Handler) http.Handler {
	static := http.FileServer(http.FS(assets))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			body, err := fs.ReadFile(assets, "index.html")
			if err != nil {
				http.Error(w, "dashboard unavailable", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/dashboard.js") || strings.HasPrefix(r.URL.Path, "/sleeper-dashboard-") {
			static.ServeHTTP(w, r)
			return
		}
		api.ServeHTTP(w, r)
	})
}
