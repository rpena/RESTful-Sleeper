package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr        string
	RedisURL        string
	SleeperBaseURL  string
	SleeperLeagueID string
	SleeperUserID   string
	RequestTimeout  time.Duration
	CacheTTL        time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr:        env("HTTP_ADDR", ":8080"),
		RedisURL:        env("REDIS_URL", "redis://redis:6379/0"),
		SleeperBaseURL:  env("SLEEPER_BASE_URL", "https://api.sleeper.app/v1"),
		SleeperLeagueID: env("SLEEPER_LEAGUE_ID", ""),
		SleeperUserID:   env("SLEEPER_USER_ID", ""),
		RequestTimeout:  durationEnv("REQUEST_TIMEOUT", 10*time.Second),
		CacheTTL:        durationEnv("CACHE_TTL", 60*time.Second),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}
