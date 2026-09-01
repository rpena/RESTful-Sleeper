package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rpena/RESTful-Sleeper/internal/cache"
	"github.com/rpena/RESTful-Sleeper/internal/sleeper"
)

type fakeClient struct {
	calls int
	path  string
}

func (c *fakeClient) Get(_ context.Context, path, _ string) (sleeper.Response, error) {
	c.calls++
	c.path = path
	return sleeper.Response{StatusCode: http.StatusOK, Body: []byte(`{"players":[]}`)}, nil
}

type fakeCache struct{ values map[string][]byte }

func (c *fakeCache) Get(_ context.Context, key string) ([]byte, error) {
	value, ok := c.values[key]
	if !ok {
		return nil, cache.ErrMiss
	}
	return value, nil
}

func TestServeUserUsesConfiguredUserID(t *testing.T) {
	client := &fakeClient{}
	store := &fakeCache{values: make(map[string][]byte)}
	h := (&Handler{sleeper: client, cache: store, ttl: time.Minute, userID: "myUser"})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/user", nil)
	recorder := httptest.NewRecorder()

	h.serve(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if client.path != "user/myUser" {
		t.Fatalf("upstream path = %q, want %q", client.path, "user/myUser")
	}
}

func TestServeUserRequiresConfiguredUserID(t *testing.T) {
	h := (&Handler{userID: ""})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/user", nil)
	recorder := httptest.NewRecorder()

	h.serve(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}
func (c *fakeCache) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	c.values[key] = value
	return nil
}

func TestServeCachesSuccessfulResponses(t *testing.T) {
	client := &fakeClient{}
	store := &fakeCache{values: make(map[string][]byte)}
	h := (&Handler{sleeper: client, cache: store, ttl: time.Minute})
	server := httptest.NewServer(http.HandlerFunc(h.serve))
	defer server.Close()

	for range 2 {
		response, err := http.Get(server.URL + "/api/v1/players/nfl")
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(response.Body)
		response.Body.Close()
		if response.StatusCode != http.StatusOK || string(body) != `{"players":[]}` {
			t.Fatalf("unexpected response: %d %s", response.StatusCode, body)
		}
	}
	if client.calls != 1 {
		t.Fatalf("upstream calls = %d, want 1", client.calls)
	}
}
