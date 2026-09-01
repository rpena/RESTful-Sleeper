package sleeper

import (
	"strings"
	"testing"
	"time"
)

func TestPlayersCachePolicyUsesUTCCalendarDay(t *testing.T) {
	now := time.Date(2026, time.September, 1, 23, 30, 0, 0, time.FixedZone("local", -7*60*60))
	key := CacheKey("players/nfl", "", now)
	if !strings.HasSuffix(key, "2026-09-02") {
		t.Fatalf("CacheKey() = %q, want UTC date suffix", key)
	}
	if ttl := CacheTTL("players/nfl", now, time.Minute); ttl <= 0 || ttl > 24*time.Hour {
		t.Fatalf("CacheTTL() = %s, want positive duration within a day", ttl)
	}
	if ttl := CacheTTL("league/test", now, time.Minute); ttl != time.Minute {
		t.Fatalf("CacheTTL() for normal endpoint = %s, want one minute", ttl)
	}
}
