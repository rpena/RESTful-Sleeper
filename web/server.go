package web

import (
	"embed"
	"net/http"
	"strings"
)

//go:embed index.html dashboard.js sleeper-dashboard-core.js sleeper-dashboard-card.js
var assets embed.FS

func NewHandler(api http.Handler) http.Handler {
	static := http.FileServer(http.FS(assets))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			r.URL.Path = "/index.html"
		}
		if strings.HasPrefix(r.URL.Path, "/index.html") || strings.HasPrefix(r.URL.Path, "/dashboard.js") || strings.HasPrefix(r.URL.Path, "/sleeper-dashboard-") {
			static.ServeHTTP(w, r)
			return
		}
		api.ServeHTTP(w, r)
	})
}
