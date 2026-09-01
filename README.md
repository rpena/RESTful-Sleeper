# RESTful Sleeper

A small Go REST API proxy for the Sleeper fantasy football API. Successful GET responses are cached in Redis with a configurable TTL, reducing repeated upstream requests and helping stay within Sleeper API limits.

## API

The proxy forwards GET requests under `/api/v1/` to Sleeper's `/v1/` API:

- `GET /api/v1/players/nfl`
- `GET /api/v1/league/{league_id}`
- `GET /api/v1/league/{league_id}/rosters`
- `GET /api/v1/league/{league_id}/matchups/{week}`
- `GET /api/v1/user/{user_id}`
- `GET /api/v1/user` (uses `SLEEPER_USER_ID`)
- `GET /api/v1/dashboard?league_id={league_id}&user_id={user_id}&season={season}&week={week}`
- `GET /healthz`

Any additional Sleeper GET endpoint can be forwarded without code changes. Query parameters are included in the Redis cache key. The `players/nfl` endpoint is the exception: it uses a UTC calendar-day key and expires at the next UTC midnight, regardless of the normal `CACHE_TTL` setting.

The dashboard endpoint returns one response containing the selected team's current matchup, starter and roster player points versus projections, league standings, and the top ten players added through completed waiver or free-agent transactions for that week. Set `SLEEPER_LEAGUE_ID` and `SLEEPER_USER_ID` in the environment for default dashboard values, or provide `league_id` and `user_id` in the request to override them. The shorthand `/api/v1/user` endpoint also uses `SLEEPER_USER_ID`. `season` defaults to 2026 and `week` defaults to 1 when omitted.

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

For Docker Compose, create a `.env` file next to `docker-compose.yml`:

```dotenv
SLEEPER_LEAGUE_ID=your-league-id
SLEEPER_USER_ID=your-sleeper-user-id
```

## Proxmox VE deployment

On the target Debian/Ubuntu VM or LXC inside Proxmox:

1. Install Docker Engine and the Compose plugin.
2. Copy this directory to the host, for example `/opt/restful-sleeper`.
3. Run `docker compose up -d --build` from that directory.
4. Restrict inbound port `8080` with the Proxmox or guest firewall, or place the API behind a reverse proxy with TLS.
5. Back up the Docker volume `redis-data` if cached data persistence is desired. The cache can also be treated as disposable and rebuilt from Sleeper.

The service listens on port `8080` and exits cleanly on SIGTERM, which supports normal container restarts and VM maintenance.
