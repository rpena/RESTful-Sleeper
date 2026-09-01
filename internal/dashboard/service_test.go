package dashboard

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/rpena/RESTful-Sleeper/internal/cache"
	"github.com/rpena/RESTful-Sleeper/internal/sleeper"
)

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name    string
		request Request
		wantErr bool
	}{
		{name: "valid", request: Request{LeagueID: "league", UserID: "user", Week: 1}},
		{name: "missing league", request: Request{UserID: "user", Week: 1}, wantErr: true},
		{name: "missing user", request: Request{LeagueID: "league", Week: 1}, wantErr: true},
		{name: "week zero", request: Request{LeagueID: "league", UserID: "user"}, wantErr: true},
		{name: "week too high", request: Request{LeagueID: "league", UserID: "user", Week: 19}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateRequest(test.request)
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateRequest() error = %v, wantErr %t", err, test.wantErr)
			}
		})
	}
}

func TestGetRefetchesMalformedCachedResponse(t *testing.T) {
	client := &testClient{body: []byte(`{"players":[]}`)}
	store := &testCache{values: map[string][]byte{"sleeper:players:nfl:2026-09-01": []byte(`{"players":`)}}
	service := New(client, store, time.Hour)

	result, err := getWithOptions[map[string]any](service, context.Background(), "players/nfl", "", "sleeper:players:nfl:2026-09-01", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if client.calls != 1 {
		t.Fatalf("upstream calls = %d, want 1", client.calls)
	}
	if _, ok := result["players"]; !ok {
		t.Fatalf("refetched result = %#v, want players field", result)
	}
}

type testClient struct {
	calls int
	body  []byte
}

func (c *testClient) Get(context.Context, string, string) (sleeper.Response, error) {
	c.calls++
	return sleeper.Response{StatusCode: http.StatusOK, Body: c.body}, nil
}

type testCache struct {
	values map[string][]byte
}

func (c *testCache) Get(_ context.Context, key string) ([]byte, error) {
	value, ok := c.values[key]
	if !ok {
		return nil, cache.ErrMiss
	}
	return value, nil
}

func (c *testCache) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	c.values[key] = value
	return nil
}
