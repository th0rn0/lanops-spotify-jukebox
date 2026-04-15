# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Layout

Two independent applications share one repo:

- `service/` — Go backend (Gin HTTP API + long-running Spotify control loop).
- `ui/` — Nuxt 3 (Vue 3) frontend that talks to the service's HTTP API.
- `resources/docker/{service,ui}/Dockerfile` — production images, built from repo root (`context: .`).
- `resources/banned-terms/{tracks.txt,words.txt}` — seed files loaded into the DB at service startup.
- `.github/workflows/docker-image.yml` — CI builds & pushes both images to Docker Hub on push to `main`.

There is no monorepo tooling; run commands from inside `service/` or `ui/`.

## Common Commands

### Service (Go)

```bash
cd service
cp .env.example .env            # fill in Spotify + admin creds before running
go mod tidy
go run ./cmd/spotify-jukebox    # entry point (NOT `go run .`)
go build ./...                  # compile check
go vet ./...
go test ./...                   # no tests exist yet, but this is the command
```

The binary requires **all** env vars in `.env.example` to be set — `config.Load()` calls `log.Fatal` on any missing var, including `BANNED_TERMS_TRACKS_FILE_LOCATION` / `BANNED_TERMS_WORDS_FILE_LOCATION` (defaults point at `../resources/banned-terms/*.txt` relative to `service/`).

### UI (Nuxt 3)

```bash
cd ui
cp .env.example .env            # set API_ENDPOINT to the service URL
npm install
npm run dev                     # http://localhost:3000
npm run build && npm run preview
```

`API_ENDPOINT` is read at build/runtime via `runtimeConfig.public.apiEndpoint` (see `nuxt.config.ts`) and baked into the UI Docker image via `--build-arg API_ENDPOINT=...`.

### Docker (from repo root)

```bash
docker build -f resources/docker/service/Dockerfile .
docker build -f resources/docker/ui/Dockerfile --build-arg API_ENDPOINT=... .
```

## Architecture

### Service: API + background loop sharing one Client

`cmd/spotify-jukebox/main.go` wires everything together:

1. Loads `.env` → `config.Config`.
2. Opens SQLite via GORM at `DB_PATH` and `AutoMigrate`s `Track`, `TrackImage`, `BannedTerm`, `AutoStart`, `oauth2.Token`.
3. Constructs one `jukebox.Client` (holds config, db, zerolog, Spotify auth+client, current track, skip/pause state).
4. Loads banned terms from the two text files into the DB (idempotent — skips existing values).
5. Starts the Gin router in a goroutine (`api.SetupRouter`) and then calls `jukeboxClient.Run()` on the main goroutine — the HTTP API and the Spotify poll loop share the **same** `*jukebox.Client` pointer, so handlers mutate the same in-memory state the loop reads.

There is no mutex around `Client` state (`active`, `paused`, `skip`, `current`). The channels package (`internal/channels`) is currently an empty placeholder — all cross-goroutine signalling happens by the handler writing to `Client` fields that the `Run()` loop polls every 5 seconds.

### The Run loop (`internal/jukebox/run.go`)

Two phases:

1. **Auth wait** — polls every 5s until a valid Spotify OAuth token is present (loaded from DB on startup, or set via `/auth/login` → `/auth/callback`). Refreshes tokens when within 30s of expiry.
2. **Main tick** (5s cadence, only when `c.active == true`):
   - If `paused`, call Spotify Pause, spin until `paused == false`, then Play.
   - Pull `PlayerState` / `GetQueue` from Spotify; if the queue has `>10` items (one real track = ~10 entries) it pauses to reclaim control from Spotify's own queueing.
   - Forces `repeat=off` and `shuffle=off` every tick.
   - If the current track finished (`!Playing || Progress == 0`) **or** `shouldSkip()` is true: add the current track to the fallback playlist (unless it was skipped or already came from fallback), delete it from the DB queue, pick next via `getNext()` (random DB row → fallback playlist), and `PlayOpt` it.

When `active == false`, the loop pauses Spotify and idles in a nested 5s sleep until re-activated.

### Track selection priority

`getNext()` does `getNextRandomFromQueue()` (`SELECT * FROM tracks ORDER BY random() LIMIT 1`); on `ErrRecordNotFound` it falls back to a random item from the configured Spotify playlist (`SPOTIFY_FALLBACK_PLAYLIST_ID`). Tracks played from the fallback playlist are flagged `FallbackPlaylist: true` (a `gorm:"-"` field, not persisted) so the loop knows not to re-add them to fallback when they end.

### Auth persistence

Spotify OAuth tokens are stored in the same SQLite DB (`oauth2.Token` table). On restart, `jukebox.New()` loads the saved token so re-auth is only needed if it's missing/expired beyond refresh. `AutoStart` table persists the `active` flag so the jukebox auto-resumes playback across restarts (see `checkForAutoStart` / `SetActive`).

### API routing (`service/api/main.go`)

- Public: `GET /tracks`, `POST /tracks/add`, `GET /tracks/current`, `GET /tracks/:trackId`, `GET /search/:searchTerm`, `POST /votes/skip`, `GET /auth/callback`.
- `gin.BasicAuth`-protected (`API_ADMIN_USERNAME` / `API_ADMIN_PASSWORD`): `GET /auth/login`, all `/player/*` endpoints.
- `POST /votes/skip` is rate-limited to 1 req per 4 minutes per `ClientIP+UserAgent` via `gin-rate-limit` in-memory store.
- CORS is wide open (`cors.Default()`).

When adding new routes: place handlers in `service/api/handlers/` (one file per resource, e.g. `tracks.go`, `player.go`), add the route in `api/main.go`, and put any new Spotify/DB logic on `*jukebox.Client` in `internal/jukebox/`.

### Banned terms

Loaded from two files at startup into a single `banned_terms` table with `type` = `"track"` (exact Spotify track ID match) or `"word"` (lowercase substring match against track name / album name / artist names). `AddTrackToQueue` calls `CheckFullTrackIsBanned` and returns `ErrTrackBanned` on hit — handlers should surface this as a user-facing rejection.

### UI

Single-page Nuxt app (`ui/app.vue`) that polls `/tracks` and `/tracks/current` every 10s. Components: `SongSearch` (search + add), `PlaylistItem` (queue row), `SearchItem` (search result row). Bootstrap 5 + SCSS (`assets/scss/main.scss`). No server routes, no Pinia — all state is local refs in `app.vue` and components, driven by `$fetch` against `runtimeConfig.public.apiEndpoint`. The "upvote" button is intentionally a joke that only changes its own label.
