package jukebox

import (
	"errors"
	"lanops/spotify-jukebox/internal/config"

	"github.com/zmb3/spotify/v2"

	"github.com/rs/zerolog"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"

	"gorm.io/gorm"
)

var (
	state = "spotifyJukeBox"
)

type Client struct {
	cfg                config.Config
	db                 *gorm.DB
	log                *zerolog.Logger
	spotify            SpotifyClient
	fallbackPlaylistId spotify.ID
	current            Track
	shuffle            bool
	active             bool
	skip               struct {
		active bool
		votes  int
	}
	paused bool
}

type SpotifyClient struct {
	auth   *spotifyauth.Authenticator
	client *spotify.Client
	token  *oauth2.Token
}

func New(cfg config.Config, db *gorm.DB, log *zerolog.Logger) (Client, error) {
	client := Client{
		cfg: cfg,
		db:  db,
		log: log,
		spotify: SpotifyClient{
			token: nil,
			auth: spotifyauth.New(
				spotifyauth.WithRedirectURL(cfg.Api.AuthCallBackUrl),
				spotifyauth.WithScopes(
					spotifyauth.ScopeUserReadCurrentlyPlaying,
					spotifyauth.ScopeUserReadPlaybackState,
					spotifyauth.ScopeUserModifyPlaybackState,
					spotifyauth.ScopePlaylistModifyPrivate,
					spotifyauth.ScopePlaylistModifyPublic,
				),
			),
		},
		current: Track{},
		shuffle: false,
		active:  false,
		paused:  false,
	}
	token, err := client.getAuthToken()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return client, err
	}

	if token != nil {
		client.setAuthToken(token, false)
	}
	client.resetSkip()
	client.checkForAutoStart()

	return client, err
}
