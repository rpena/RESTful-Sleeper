package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/rpena/RESTful-Sleeper/internal/cache"
	"github.com/rpena/RESTful-Sleeper/internal/sleeper"
)

type sleeperClient interface {
	Get(context.Context, string, string) (sleeper.Response, error)
}

type cacheStore interface {
	Get(context.Context, string) ([]byte, error)
	Set(context.Context, string, []byte, time.Duration) error
}

type Handler struct {
	sleeper sleeperClient
	cache   cacheStore
	ttl     time.Duration
}

func New(client *sleeper.Client, redisCache *cache.Cache, ttl time.Duration) http.Handler {
	h := &Handler{sleeper: client, cache: redisCache, ttl: ttl}
	return http.HandlerFunc(h.serve)
}

func (h *Handler) serve(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
		return
	}
	if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, "/api/v1/") {
		http.NotFound(w, r)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/")
	key := "sleeper:" + r.URL.RequestURI()
	if body, err := h.cache.Get(r.Context(), key); err == nil {
		writeJSON(w, http.StatusOK, body)
		return
	} else if err != cache.ErrMiss {
		slog.Debug("cache read failed", "error", err, "key", key)
	}

	response, err := h.sleeper.Get(r.Context(), path, r.URL.RawQuery)
	if err != nil {
		http.Error(w, `{"error":"upstream unavailable"}`, http.StatusBadGateway)
		return
	}
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		if err := h.cache.Set(r.Context(), key, response.Body, h.ttl); err != nil {
			slog.Debug("cache write failed", "error", err, "key", key)
		}
	}
	writeJSON(w, response.StatusCode, response.Body)
}

func writeJSON(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
