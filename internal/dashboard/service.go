package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rpena/RESTful-Sleeper/internal/cache"
	"github.com/rpena/RESTful-Sleeper/internal/sleeper"
)

type Service struct {
	client sleeperClient
	cache  cacheStore
	ttl    time.Duration
}

type sleeperClient interface {
	Get(context.Context, string, string) (sleeper.Response, error)
}

type cacheStore interface {
	Get(context.Context, string) ([]byte, error)
	Set(context.Context, string, []byte, time.Duration) error
}

type Request struct {
	LeagueID string
	UserID   string
	Season   string
	Week     int
}

type Dashboard struct {
	League        League         `json:"league"`
	Week          int            `json:"week"`
	Matchup       MatchupView    `json:"matchup"`
	Standings     []Standing     `json:"standings"`
	WaiverPickups []WaiverPickup `json:"waiver_pickups"`
}

type League struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Season       string `json:"season"`
	TotalRosters int    `json:"total_rosters"`
}

type MatchupView struct {
	Opponent *TeamMatchup `json:"opponent,omitempty"`
	YourTeam *TeamMatchup `json:"your_team,omitempty"`
}

type TeamMatchup struct {
	RosterID int      `json:"roster_id"`
	OwnerID  string   `json:"owner_id,omitempty"`
	Points   float64  `json:"points"`
	Players  []Player `json:"players"`
}

type Player struct {
	ID        string  `json:"id"`
	FullName  string  `json:"full_name,omitempty"`
	Position  string  `json:"position,omitempty"`
	Starter   bool    `json:"starter"`
	Projected float64 `json:"projected_points"`
	Points    float64 `json:"points"`
}

type Standing struct {
	Rank        int     `json:"rank"`
	RosterID    int     `json:"roster_id"`
	OwnerID     string  `json:"owner_id,omitempty"`
	DisplayName string  `json:"display_name,omitempty"`
	Wins        int     `json:"wins"`
	Losses      int     `json:"losses"`
	Ties        int     `json:"ties"`
	PointsFor   float64 `json:"points_for"`
}

type WaiverPickup struct {
	PlayerID string `json:"player_id"`
	Name     string `json:"name,omitempty"`
	Position string `json:"position,omitempty"`
	Adds     int    `json:"adds"`
}

type leagueResponse struct {
	Name         string `json:"name"`
	Season       string `json:"season"`
	TotalRosters int    `json:"total_rosters"`
}

type userResponse struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}

type rosterResponse struct {
	RosterID int      `json:"roster_id"`
	OwnerID  string   `json:"owner_id"`
	Players  []string `json:"players"`
	Starters []string `json:"starters"`
	Settings struct {
		Wins      int     `json:"wins"`
		Losses    int     `json:"losses"`
		Ties      int     `json:"ties"`
		PointsFor float64 `json:"fpts"`
	} `json:"settings"`
}

type matchupResponse struct {
	RosterID  int     `json:"roster_id"`
	MatchupID int     `json:"matchup_id"`
	Points    float64 `json:"points"`
}

type transactionResponse struct {
	Type   string            `json:"type"`
	Status string            `json:"status"`
	Adds   map[string]string `json:"adds"`
}

type playerMetadata struct {
	FullName string `json:"full_name"`
	Position string `json:"position"`
}

func New(client sleeperClient, redisCache cacheStore, ttl time.Duration) *Service {
	return &Service{client: client, cache: redisCache, ttl: ttl}
}

func (s *Service) Build(ctx context.Context, request Request) (Dashboard, error) {
	league, err := get[leagueResponse](s, ctx, "league/"+request.LeagueID, "")
	if err != nil {
		return Dashboard{}, err
	}
	if request.Season == "" {
		request.Season = league.Season
	}
	users, _ := get[[]userResponse](s, ctx, "league/"+request.LeagueID+"/users", "")
	rosters, err := get[[]rosterResponse](s, ctx, "league/"+request.LeagueID+"/rosters", "")
	if err != nil {
		return Dashboard{}, err
	}
	matchups, err := get[[]matchupResponse](s, ctx, "league/"+request.LeagueID+"/matchups/"+strconv.Itoa(request.Week), "")
	if err != nil {
		return Dashboard{}, err
	}
	stats, _ := get[map[string]map[string]any](s, ctx, "stats/nfl/"+request.Season+"/"+strconv.Itoa(request.Week), "season_type=regular")
	projections, _ := get[map[string]map[string]any](s, ctx, "projections/nfl/"+request.Season+"/"+strconv.Itoa(request.Week), "season_type=regular")
	now := time.Now().UTC()
	players, err := getWithOptions[map[string]playerMetadata](s, ctx, "players/nfl", "", sleeper.CacheKey("players/nfl", "", now), sleeper.CacheTTL("players/nfl", now, s.ttl))
	if err != nil {
		return Dashboard{}, err
	}

	ownerNames := make(map[string]string, len(users))
	for _, user := range users {
		ownerNames[user.UserID] = user.DisplayName
	}
	rosterByID := make(map[int]rosterResponse, len(rosters))
	for _, roster := range rosters {
		rosterByID[roster.RosterID] = roster
	}
	matchupByRoster := make(map[int]matchupResponse, len(matchups))
	for _, matchup := range matchups {
		matchupByRoster[matchup.RosterID] = matchup
	}

	standings := make([]Standing, 0, len(rosters))
	for _, roster := range rosters {
		standings = append(standings, Standing{RosterID: roster.RosterID, OwnerID: roster.OwnerID, DisplayName: ownerNames[roster.OwnerID], Wins: roster.Settings.Wins, Losses: roster.Settings.Losses, Ties: roster.Settings.Ties, PointsFor: roster.Settings.PointsFor})
	}
	sort.SliceStable(standings, func(i, j int) bool { return standings[i].PointsFor > standings[j].PointsFor })
	for index := range standings {
		standings[index].Rank = index + 1
	}

	var matchupView MatchupView
	for _, roster := range rosters {
		if roster.OwnerID != request.UserID {
			continue
		}
		matchup, ok := matchupByRoster[roster.RosterID]
		if !ok {
			break
		}
		matchupView.YourTeam = s.teamMatchup(roster, matchup, players, stats, projections)
		for _, opponent := range matchups {
			if opponent.MatchupID == matchup.MatchupID && opponent.RosterID != roster.RosterID {
				if opponentRoster, exists := rosterByID[opponent.RosterID]; exists {
					matchupView.Opponent = s.teamMatchup(opponentRoster, opponent, players, stats, projections)
					break
				}
			}
		}
		break
	}

	waiverPickups := s.waiverPickups(ctx, request.LeagueID, request.Week, players)
	return Dashboard{League: League{ID: request.LeagueID, Name: league.Name, Season: league.Season, TotalRosters: league.TotalRosters}, Week: request.Week, Matchup: matchupView, Standings: standings, WaiverPickups: waiverPickups}, nil
}

func (s *Service) teamMatchup(roster rosterResponse, matchup matchupResponse, players map[string]playerMetadata, stats, projections map[string]map[string]any) *TeamMatchup {
	team := &TeamMatchup{RosterID: roster.RosterID, OwnerID: roster.OwnerID, Points: matchup.Points, Players: make([]Player, 0, len(roster.Players))}
	starters := make(map[string]bool, len(roster.Starters))
	for _, id := range roster.Starters {
		starters[id] = true
	}
	for _, id := range roster.Players {
		metadata := players[id]
		team.Players = append(team.Players, Player{ID: id, FullName: metadata.FullName, Position: metadata.Position, Starter: starters[id], Projected: metric(projections[id], "pts_ppr"), Points: metric(stats[id], "pts_ppr")})
	}
	sort.SliceStable(team.Players, func(i, j int) bool { return team.Players[i].Starter && !team.Players[j].Starter })
	return team
}

func metric(values map[string]any, key string) float64 {
	value, ok := values[key]
	if !ok {
		return 0
	}
	switch number := value.(type) {
	case float64:
		return number
	case json.Number:
		parsed, _ := number.Float64()
		return parsed
	default:
		return 0
	}
}

func (s *Service) waiverPickups(ctx context.Context, leagueID string, week int, players map[string]playerMetadata) []WaiverPickup {
	transactions, err := get[[]transactionResponse](s, ctx, "league/"+leagueID+"/transactions/"+strconv.Itoa(week), "")
	if err != nil {
		return []WaiverPickup{}
	}
	counts := make(map[string]int)
	for _, transaction := range transactions {
		if transaction.Status != "complete" || (transaction.Type != "waiver" && transaction.Type != "free_agent") {
			continue
		}
		for playerID := range transaction.Adds {
			counts[playerID]++
		}
	}
	pickups := make([]WaiverPickup, 0, len(counts))
	for playerID, adds := range counts {
		metadata := players[playerID]
		pickups = append(pickups, WaiverPickup{PlayerID: playerID, Name: metadata.FullName, Position: metadata.Position, Adds: adds})
	}
	sort.SliceStable(pickups, func(i, j int) bool { return pickups[i].Adds > pickups[j].Adds })
	if len(pickups) > 10 {
		pickups = pickups[:10]
	}
	return pickups
}

func get[T any](s *Service, ctx context.Context, path, query string) (T, error) {
	return getWithOptions[T](s, ctx, path, query, "", s.ttl)
}

func getWithOptions[T any](s *Service, ctx context.Context, path, query, cacheKey string, ttl time.Duration) (T, error) {
	var result T
	cacheHit := false
	if cacheKey == "" {
		cacheKey = "sleeper:" + path
		if query != "" {
			cacheKey += "?" + query
		}
	}
	body, err := s.cache.Get(ctx, cacheKey)
	if err == cache.ErrMiss {
		response, requestErr := s.client.Get(ctx, path, query)
		if requestErr != nil {
			return result, requestErr
		}
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return result, fmt.Errorf("Sleeper returned status %d for %s", response.StatusCode, path)
		}
		body = response.Body
		if cacheErr := s.cache.Set(ctx, cacheKey, body, ttl); cacheErr != nil { /* cache failures do not block a dashboard response */
		}
	} else if err != nil {
		return result, fmt.Errorf("read cache for %s: %w", path, err)
	} else {
		cacheHit = true
	}
	if err := json.Unmarshal(body, &result); err != nil {
		if !cacheHit {
			return result, fmt.Errorf("decode Sleeper response for %s: %w", path, err)
		}
		response, requestErr := s.client.Get(ctx, path, query)
		if requestErr != nil {
			return result, requestErr
		}
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return result, fmt.Errorf("Sleeper returned status %d for %s", response.StatusCode, path)
		}
		if cacheErr := s.cache.Set(ctx, cacheKey, response.Body, ttl); cacheErr != nil { /* cache failures do not block a dashboard response */
		}
		if err := json.Unmarshal(response.Body, &result); err != nil {
			return result, fmt.Errorf("decode Sleeper response for %s: %w", path, err)
		}
	}
	return result, nil
}

func ValidateRequest(request Request) error {
	if strings.TrimSpace(request.LeagueID) == "" {
		return fmt.Errorf("league_id is required")
	}
	if strings.TrimSpace(request.UserID) == "" {
		return fmt.Errorf("user_id is required")
	}
	if request.Week < 1 || request.Week > 18 {
		return fmt.Errorf("week must be between 1 and 18")
	}
	return nil
}
