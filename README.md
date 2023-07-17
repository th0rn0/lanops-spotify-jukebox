# Spotify Jukebox

## API

Some installation instructions.

- https://developer.spotify.com/documentation/web-api/concepts/apps - Create App for Client and Secret Key

```nodemon --exec go run main.go --signal SIGTERM```

### Limitations

Because of how the api polls for the progress of the current track, we cant just give spotify a full playlist to play, instead the API will pick songs at random should a playlist be given to play

## UI

Some installation instructions.



# Suggestions

## API
- Add Queued Songs to Fallback Playlist
- Admin Only Access/Function
- JWT Tokens for Admin Access
- Better Auth Flow
- Get current song brings back streamlined request with progress (for frontend display)

