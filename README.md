# RESTful Sleeper

A small Go REST API proxy for the Sleeper fantasy football API. Successful GET responses are cached in Redis with a configurable TTL, reducing repeated upstream requests and helping stay within Sleeper API limits.

## API

The proxy forwards GET requests under `/api/v1/` to Sleeper's `/v1/` API:

- `GET /api/v1/players/nfl`
- `GET /api/v1/leagues/{league_id}`
- `GET /api/v1/league/{league_id}/rosters`
- `GET /api/v1/league/{league_id}/matchups/{week}`
- `GET /api/v1/user/{user_id}`
- `GET /healthz`

Any additional Sleeper GET endpoint can be forwarded without code changes. Query parameters are included in the Redis cache key.

## Local development

Requirements: Go 1.27+ and Docker Compose.

```sh
go mod tidy
go test ./...
docker compose up -d --build
curl http://localhost:8080/api/v1/players/nfl
```

Stop the stack with `docker compose down`. Redis data is kept in the `redis-data` named volume.

`CACHE_TTL` and `REQUEST_TIMEOUT` are expressed in seconds. Override them in a local `.env` file or in the Compose environment section.

## Proxmox VE deployment

On the target Debian/Ubuntu VM or LXC inside Proxmox:

1. Install Docker Engine and the Compose plugin.
2. Copy this directory to the host, for example `/opt/restful-sleeper`.
3. Run `docker compose up -d --build` from that directory.
4. Restrict inbound port `8080` with the Proxmox or guest firewall, or place the API behind a reverse proxy with TLS.
5. Back up the Docker volume `redis-data` if cached data persistence is desired. The cache can also be treated as disposable and rebuilt from Sleeper.

The service listens on port `8080` and exits cleanly on SIGTERM, which supports normal container restarts and VM maintenance.
