# Spotify Jukebox

## API

### Prerequisites

- Create App for Client and Secret Key. Make sure to set the callback url to the domain and have the callback path, for example ```http://localhost:8888/auth/callback```
- - https://developer.spotify.com/documentation/web-api/concepts/apps -
- Copy the example env file ```cp .env.example .env```
- Fill in the ```.env``` file
- - Set Device ID only if known
- - Fallback Playlist can be any playlist. Use the full URI. If you wish to add queued songs to the playlist make sure the account being used has the sufficient permissions to the playlist

### Install Dependencies
```bash
    cd api
    go mod tidy
```

### Run

```bash
    go run .
```

### Usage

To initate the player, you must first request the login link from the ```/auth/login``` endpoint. Then go to the returned link and log into Spotify. Once this has completed you must use the ```localhost:8888/device``` endpoint to set the Device.

To start the player goto the ```/player/start``` endpoint

Admin routes are behing Basic Auth Endpoints. The User is ```admin``` and the password is set in the env.

## UI

Some installation instructions.



# Suggestions

## API

## UI

