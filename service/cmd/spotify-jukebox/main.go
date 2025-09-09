package main

import (
	"errors"
	"fmt"
	"lanops/spotify-jukebox/api"
	"lanops/spotify-jukebox/internal/channels"
	"lanops/spotify-jukebox/internal/config"
	"lanops/spotify-jukebox/internal/jukebox"
	"os"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	logger zerolog.Logger
	cfg    config.Config
	msgCh  = make(chan channels.MsgCh, 20)
	db     *gorm.DB
)

func main() {
	logger = zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
	).Level(zerolog.TraceLevel).With().Timestamp().Caller().Logger()
	logger.Info().Msg("Initializing Spotify Jukebox")

	logger.Info().Msg("Loading Config")
	cfg = config.Load()

	// Load Database & Migrate the schema
	db, err := gorm.Open(sqlite.Open(cfg.DbPath), &gorm.Config{})
	logger.Info().Msg("Connecting to Database")
	if err != nil {
		logger.Fatal().Err(err).Msg("Error Connecting to Database")
	}

	db.AutoMigrate(&jukebox.Track{})
	db.AutoMigrate(&jukebox.TrackImage{})
	db.AutoMigrate(&jukebox.BannedTerm{})
	db.AutoMigrate(&oauth2.Token{})

	jukeboxClient, err := jukebox.New(cfg, db, &logger)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Fatal().Err(err).Msg("Getting Jukebox Client")
	}

	logger.Info().Msg(fmt.Sprintf("Setting Fallback playlist to %s", cfg.Spotify.FallbackPlaylistId))
	jukeboxClient.SetFallbackPlaylist(cfg.Spotify.FallbackPlaylistId)

	logger.Info().Msg("Loading Banned Terms from file")
	if err = jukeboxClient.LoadBannedTermsFromFile(); err != nil {
		logger.Fatal().Err(err).Msg("Could not load Banned Terms from file")
	}

	logger.Info().Msg("Starting Spotify Jukebox API")
	api := api.SetupRouter(cfg, &jukeboxClient, logger)
	go func() {
		api.Run(fmt.Sprintf(":%s", cfg.Api.Port))
	}()

	logger.Info().Msg("Starting Spotify Jukebox Client")
	if err := jukeboxClient.Run(); err != nil {
		logger.Fatal().Err(err).Msg("Something went wrong with the jukebox")
	}
}
