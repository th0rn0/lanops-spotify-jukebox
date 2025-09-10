# LanOps Spotify Jukebox

Spotify Jukebox System written in GO and JS. Once logged in, the jukebox will hook onto the active spotify device and play songs from a queue. Songs are then added to a fallback playlist for use later or when the jukebox has no songs in the queue.

### Features

- Add/Remove/Skip Tracks
- Voting system - Vote To skip songs
- Admin Controls
- Volume Controls
- Works with any Spotify Device
- Fallback Playlist when no songs in queue
- Add Queued songs to Fallback Playlist

Intended for use at [LanOps Events](https://www.lanops.co.uk)

Thanks to everyone involved in the following projects:

https://github.com/zmb3/spotify

## Service

Written in GO using the https://github.com/zmb3/spotify package. Refer to the postman collection for the endpoints

### Prerequisites

- Create Spotify App for Client and Secret Key. Make sure to set the callback url to the domain and have the callback path, for example ```http://localhost:8888/auth/callback```
    - https://developer.spotify.com/documentation/web-api/concepts/apps 
- ```cp service/.env.example service/.env``` and fill it in
    - Fallback Playlist can be any playlist. Use the full URI. If you wish to add queued songs to the playlist make sure the account being used has the sufficient permissions to the playlist.

#### Install Dependencies

```bash
    cd service
    go mod tidy
```

### Usage

Entry Point:
```bash
    go run ./cmd/spotify-jukebox
```

### API Endpoints

| Endpoint              | Method | URL Params          | JSON Input          | Description                                   |
|-----------------------|--------|---------------------|---------------------|-----------------------------------------------|
| `/votes/skip`         | POST   | None                | None                | Cast a vote to skip the current track.        |
| `/tracks`             | GET    | None                | None                | Retrieve a list of all tracks.                |
| `/tracks/add`         | POST   | None                | None                | Add a new track to the playlist.              |
| `/tracks/current`     | GET    | None                | None                | Get the currently playing track.              |
| `/tracks/:trackId`    | GET    | `trackId` (path)    | None                | Get details of a specific track by ID.        |
| `/search/:searchTerm` | GET    | `searchTerm` (path) | None                | Search tracks by a term.                      |
| `/auth/callback`      | GET    | None                | None                | OAuth callback endpoint after authentication. |
| `/auth/login`         | GET    | None                | None                | Initiate login process (requires auth).       |
| `/player/start`       | POST   | None                | None                | Start playback of the player.                 |
| `/player/stop`        | POST   | None                | None                | Stop playback of the player.                  |
| `/player/volume`      | POST   | None                | `{ "volume": int }` | Set player volume.                            |
| `/player/skip`        | POST   | None                | None                | Skip the currently playing track.             |
| `/player/pause`       | POST   | None                | None                | Pause the currently playing track.            |

### Env

| Variable                            | Description                                                      |
|-------------------------------------|------------------------------------------------------------------|
| `DB_PATH`                           | Path to the jukebox SQLite database file.                        |
| `VOTE_COUNT_TO_SKIP`                | Number of votes required to skip the current track.              |
| `SPOTIFY_ID`                        | Spotify API client ID.                                           |
| `SPOTIFY_SECRET`                    | Spotify API client secret.                                       |
| `SPOTIFY_FALLBACK_PLAYLIST_ID`      | Spotify playlist ID to use as fallback if no track is available. |
| `API_AUTH_CALLBACK_URL`             | URL for the API OAuth callback endpoint.                         |
| `API_ADMIN_PASSWORD`                | Password for the API admin account.                              |
| `API_ADMIN_USERNAME`                | Username for the API admin account.                              |
| `API_PORT`                          | Port the API server listens on.                                  |
| `BANNED_TERMS_TRACKS_FILE_LOCATION` | Path to the file containing banned track names.                  |
| `BANNED_TERMS_WORDS_FILE_LOCATION`  | Path to the file containing banned words.                        |

### Docker

```docker build -f resources/docker/service/Dockerfile .```
```
docker run -d \
  --name jukebox-service \
  --restart unless-stopped \
  -e DB_PATH=/db/jukebox.db \
  -e VOTE_COUNT_TO_SKIP=5 \
  -e SPOTIFY_ID= \
  -e SPOTIFY_SECRET= \
  -e SPOTIFY_FALLBACK_PLAYLIST_ID= \
  -e API_AUTH_CALLBACK_URL= \
  -e API_ADMIN_USERNAME= \
  -e API_ADMIN_PASSWORD= \
  -e API_PORT=20 \
  -e BANNED_TERMS_TRACKS_FILE_LOCATION= \
  -e BANNED_TERMS_WORDS_FILE_LOCATION= \
  -p 8888:8888 \
  -v /mnt/servdata/lanops/jukebox/db:/db \
  th0rn0/lanops-spotify-jukebox:service-latest
```

```
  jukebox-service:
    image: th0rn0/lanops-spotify-jukebox:service-latest
    restart: unless-stopped
    environment:
      - DB_PATH=/db/jukebox.db
      - VOTE_COUNT_TO_SKIP=5
      - SPOTIFY_ID=
      - SPOTIFY_SECRET=
      - SPOTIFY_FALLBACK_PLAYLIST_ID=
      - API_AUTH_CALLBACK_URL=
      - API_ADMIN_USERNAME=
      - API_ADMIN_PASSWORD=
      - API_PORT=20
      - BANNED_TERMS_TRACKS_FILE_LOCATION=
      - BANNED_TERMS_WORDS_FILE_LOCATION=
    ports:
      - 8888:8888
    volumes:
      - /mnt/servdata/lanops/jukebox/db:/db
```

## Frontend

### Prerequisites
- ```cp ui/.env.example ui/.env``` and fill it in

#### Install Dependencies
```bash
    cd ui
    npm install
```

### Run

```bash
    cd ui
    npm run dev
```

### Env

| Variable                            | Description                                                      |
|-------------------------------------|------------------------------------------------------------------|
| `API_ENDPOINT`                      | URL of the API.                                                  |

### Docker

```docker build -f resources/docker/ui/Dockerfile .```

```
docker run -d \
  --name jukebox-ui \
  --restart unless-stopped \
  th0rn0/lanops-spotify-jukebox:ui-latest
```

```
  jukebox-ui:
    image: th0rn0/lanops-spotify-jukebox:ui-latest
    restart: unless-stopped
    ports:
```