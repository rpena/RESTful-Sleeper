package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rpena/RESTful-Sleeper/internal/cache"
	"github.com/rpena/RESTful-Sleeper/internal/dashboard"
	"github.com/rpena/RESTful-Sleeper/internal/sleeper"
)

type sleeperClient interface {
	Get(context.Context, string, string) (sleeper.Response, error)
}

type cacheStore interface {
	Get(context.Context, string) ([]byte, error)
	Set(context.Context, string, []byte, time.Duration) error
}

type dashboardService interface {
	Build(context.Context, dashboard.Request) (dashboard.Dashboard, error)
}

type Handler struct {
	sleeper   sleeperClient
	cache     cacheStore
	ttl       time.Duration
	dashboard dashboardService
}

func New(client *sleeper.Client, redisCache *cache.Cache, ttl time.Duration, dashboardServices ...dashboardService) http.Handler {
	var dashboardHandler dashboardService
	if len(dashboardServices) > 0 {
		dashboardHandler = dashboardServices[0]
	}
	h := &Handler{sleeper: client, cache: redisCache, ttl: ttl, dashboard: dashboardHandler}
	return http.HandlerFunc(h.serve)
}

func (h *Handler) serve(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
		return
	}
	if r.URL.Path == "/api/v1/dashboard" {
		h.serveDashboard(w, r)
		return
	}
	if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, "/api/v1/") {
		http.NotFound(w, r)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/")
	key := sleeper.CacheKey(path, r.URL.RawQuery, time.Now().UTC())
	ttl := sleeper.CacheTTL(path, time.Now().UTC(), h.ttl)
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
		if err := h.cache.Set(r.Context(), key, response.Body, ttl); err != nil {
			slog.Debug("cache write failed", "error", err, "key", key)
		}
	}
	writeJSON(w, response.StatusCode, response.Body)
}

func (h *Handler) serveDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || h.dashboard == nil {
		http.NotFound(w, r)
		return
	}
	week, err := strconv.Atoi(r.URL.Query().Get("week"))
	if err != nil {
		week = 1
	}
	request := dashboard.Request{
		LeagueID: r.URL.Query().Get("league_id"),
		UserID:   r.URL.Query().Get("user_id"),
		Season:   r.URL.Query().Get("season"),
		Week:     week,
	}
	if err := dashboard.ValidateRequest(request); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	result, err := h.dashboard.Build(r.Context(), request)
	if err != nil {
		slog.Error("dashboard build failed", "error", err)
		http.Error(w, `{"error":"dashboard unavailable"}`, http.StatusBadGateway)
		return
	}
	body, err := json.Marshal(result)
	if err != nil {
		http.Error(w, `{"error":"dashboard encoding failed"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, body)
}

func writeJSON(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
