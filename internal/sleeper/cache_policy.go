package sleeper

import (
	"strings"
	"time"
)

func CacheKey(path, query string, now time.Time) string {
	if strings.Trim(path, "/") == "players/nfl" {
		return "sleeper:players:nfl:" + now.UTC().Format("2006-01-02")
	}
	key := "sleeper:" + path
	if query != "" {
		key += "?" + query
	}
	return key
}

func CacheTTL(path string, now time.Time, defaultTTL time.Duration) time.Duration {
	if strings.Trim(path, "/") != "players/nfl" {
		return defaultTTL
	}
	utcNow := now.UTC()
	nextDay := time.Date(utcNow.Year(), utcNow.Month(), utcNow.Day()+1, 0, 0, 0, 0, time.UTC)
	ttl := nextDay.Sub(utcNow)
	if ttl < time.Second {
		return time.Second
	}
	return ttl
}
