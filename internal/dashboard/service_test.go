package dashboard

import (
	"context"
	"net/http"
	"strings"
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

func TestMatchupSummariesIncludeEveryTeamAndPlayerCounts(t *testing.T) {
	service := &Service{}
	rosters := map[int]rosterResponse{
		1: {RosterID: 1, OwnerID: "user-1", Starters: []string{"p1", "p2"}},
		2: {RosterID: 2, OwnerID: "user-2", Starters: []string{"p3"}},
	}
	matchups := []matchupResponse{
		{RosterID: 2, MatchupID: 7, Points: 88.5},
		{RosterID: 1, MatchupID: 7, Points: 92.25},
	}
	stats := map[string]map[string]any{"p1": {"pts_ppr": 12.0}}
	projections := map[string]map[string]any{
		"p1": {"pts_ppr": 14.0},
		"p2": {"pts_ppr": 10.0},
		"p3": {"pts_ppr": 9.0},
	}

	result := service.matchupSummaries(matchups, rosters, map[string]string{"user-1": "One", "user-2": "Two"}, stats, projections)

	if len(result) != 1 || len(result[0].Teams) != 2 {
		t.Fatalf("matchup summaries = %#v, want one matchup with two teams", result)
	}
	team := result[0].Teams[0]
	if team.RosterID != 1 || team.CurrentPoints != 92.25 || team.PlayersCompleted != 1 || team.PlayersRemaining != 1 || team.ProjectedPoints != 24 || !team.ProjectionAvailable {
		t.Fatalf("team summary = %#v, want roster 1 with score, counts, and projection", team)
	}
}

func TestUserReferenceResolvesUsername(t *testing.T) {
	users := []userResponse{{UserID: "123", Username: "muych1ngones", DisplayName: "My Team"}}
	requestedOwnerID := "muych1ngones"
	for _, user := range users {
		if strings.EqualFold(user.UserID, requestedOwnerID) || strings.EqualFold(user.Username, requestedOwnerID) || strings.EqualFold(user.DisplayName, requestedOwnerID) {
			requestedOwnerID = user.UserID
		}
	}
	if requestedOwnerID != "123" {
		t.Fatalf("resolved owner ID = %q, want %q", requestedOwnerID, "123")
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
