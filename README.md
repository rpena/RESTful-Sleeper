# RESTful Sleeper

A small Go REST API proxy for the Sleeper fantasy football API. Successful GET responses are cached in Redis with a configurable TTL, reducing repeated upstream requests and helping stay within Sleeper API limits.

## API

The proxy forwards GET requests under `/api/v1/` to Sleeper's `/v1/` API:

- `GET /api/v1/players/nfl`
- `GET /api/v1/league/{league_id}`
- `GET /api/v1/league/{league_id}/rosters`
- `GET /api/v1/league/{league_id}/matchups/{week}`
- `GET /api/v1/league` (uses `SLEEPER_LEAGUE_ID`)
- `GET /api/v1/league/rosters` (uses `SLEEPER_LEAGUE_ID`)
- `GET /api/v1/league/matchups/{week}` (uses `SLEEPER_LEAGUE_ID`)
- `GET /api/v1/user/{user_id}`
- `GET /api/v1/user` (uses `SLEEPER_USER_ID`)
- `GET /api/v1/dashboard?league_id={league_id}&user_id={user_id}&season={season}&week={week}`
- `GET /healthz`

Any additional Sleeper GET endpoint can be forwarded without code changes. Query parameters are included in the Redis cache key. The `players/nfl` endpoint is the exception: it uses a UTC calendar-day key and expires at the next UTC midnight, regardless of the normal `CACHE_TTL` setting.

The dashboard endpoint returns one response containing every matchup for the selected week, with current scores, projected scores when available, and starter counts split between completed and remaining players. It also includes the selected team's matchup, player points versus projections, league standings, and the top ten players added through completed waiver or free-agent transactions. Set `SLEEPER_LEAGUE_ID` and `SLEEPER_USER_ID` in the environment for default dashboard values, or provide `league_id` and `user_id` in the request to override them. The shorthand `/api/v1/user` endpoint also uses `SLEEPER_USER_ID`. The shorthand league endpoints use `SLEEPER_LEAGUE_ID`; explicit IDs in the URL remain supported. `season` defaults to the configured league's season and `week` defaults to 1 when omitted. Users, stats, projections, and waiver data are supplemental; if one is unavailable, the dashboard still returns its available sections.

Dashboard refreshes are gated to NFL game windows (Thursday, Sunday, and Monday during September through February). On those days, the service caches the Sleeper `state/nfl` response until the next UTC midnight before fetching dashboard data. On other days, it returns `game_day: false` without calling Sleeper, so normal cache TTL expiry does not trigger unnecessary upstream requests. This weekday gate is intentionally conservative because Sleeper's state endpoint does not expose an exact daily game schedule.

## Local development

Requirements: Go 1.27+ and Docker Compose.

```sh
go mod tidy
go test ./...
docker compose up -d --build
curl http://localhost:8080/api/v1/players/nfl
```

Stop the stack with `docker compose down`. Redis data is kept in the `redis-data` named volume, so restarting the stack on the same UTC day reuses the cached players response. Do not use `docker compose down -v` unless you intentionally want to delete the cache and allow another players request.

`CACHE_TTL` and `REQUEST_TIMEOUT` are expressed in seconds. Override them in a local `.env` file or in the Compose environment section.

## Cache backends

Redis is the default cache backend and is included in Docker Compose. The dashboard and HTTP handler depend on the `cache.Store` interface rather than Redis directly, so another backend can be added without changing those layers. A replacement only needs to provide `Get` and `Set` methods:

```go
type Store interface {
  Get(context.Context, string) ([]byte, error)
  Set(context.Context, string, []byte, time.Duration) error
}
```

To use a different backend, implement this interface and inject it when constructing the dashboard service and HTTP handler in `cmd/api/main.go`. Redis-specific setup, such as the Compose service and startup ping, can then be replaced by that backend's initialization and health check.

For Docker Compose, create a `.env` file next to `docker-compose.yml`:

```dotenv
SLEEPER_LEAGUE_ID=your-league-id
SLEEPER_USER_ID=your-sleeper-username-or-user-id
```

## Proxmox VE deployment

On the target Debian/Ubuntu VM or LXC inside Proxmox:

1. Install Docker Engine and the Compose plugin.
2. Copy this directory to the host, for example `/opt/restful-sleeper`.
3. Run `docker compose up -d --build` from that directory.
4. Restrict inbound port `8080` with the Proxmox or guest firewall, or place the API behind a reverse proxy with TLS.
5. Back up the Docker volume `redis-data` if cached data persistence is desired. The cache can also be treated as disposable and rebuilt from Sleeper.

The service listens on port `8080` and exits cleanly on SIGTERM, which supports normal container restarts and VM maintenance.

The Go service also serves the standalone browser dashboard at `http://YOUR_DOCKER_HOST:8080/`. It uses the same `web/sleeper-dashboard-core.js` module as the Home Assistant card, so both clients share formatting and visual styles. The page defaults to `/api/v1/dashboard`; use its endpoint field if the API is hosted elsewhere.

## Home Assistant card

The repository includes a dependency-free custom Lovelace card at `web/sleeper-dashboard-card.js`. Copy both `web/sleeper-dashboard-card.js` and `web/sleeper-dashboard-core.js` into Home Assistant's `/config/www/` directory, then add the card as a JavaScript module resource under **Settings > Dashboards > Resources**:

```yaml
url: /local/sleeper-dashboard-card.js
type: module
```

Configure the REST sensor first:

```yaml
rest:
  - resource: http://YOUR_DOCKER_HOST:8080/api/v1/dashboard
    scan_interval: 60
    sensor:
      - name: Sleeper Dashboard
        unique_id: sleeper_dashboard
        value_template: "{{ value_json.week }}"
        json_attributes:
          - league
          - game_day
          - status
          - matchups
          - matchup
          - standings
          - waiver_pickups
```

Add the card to a dashboard:

```yaml
type: custom:sleeper-dashboard-card
entity: sensor.sleeper_dashboard
title: Gameday board
show_standings: true
show_waivers: true
show_raw: false
```

The refresh button requests an immediate update of the REST sensor. The API remains available as raw JSON at `/api/v1/dashboard`.

The REST sensor polls automatically, so no `shell_command` or update automation is required. If you have an automation named `Update Sleeper Users`, remove its `shell_command.update_sleeper_users` action. If you want a scheduled refresh instead of relying only on `scan_interval`, use the built-in service:

```yaml
alias: Update Sleeper Dashboard
trigger:
	- platform: time_pattern
		minutes: "/5"
action:
	- service: homeassistant.update_entity
		target:
			entity_id: sensor.sleeper_dashboard
mode: single
```
