package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify/v2"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	deviceID         spotify.ID
	db               *gorm.DB
	auth             *spotifyauth.Authenticator
	minimumVotes     int64
	fallbackPlaylist FallbackPlaylist

	// ch    = make(chan *spotify.Client)
	state = "spotifyJukeBox"
)

func main() {
	// Load Env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Load Spotify API
	auth = spotifyauth.New(
		spotifyauth.WithRedirectURL(os.Getenv("CALLBACK_URL")),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadCurrentlyPlaying, spotifyauth.ScopeUserReadPlaybackState, spotifyauth.ScopeUserModifyPlaybackState),
	)

	// Load Database
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// Migrate the schema
	db.AutoMigrate(&Track{})
	db.AutoMigrate(&TrackImage{})

	// Set Device ID
	deviceID = spotify.ID(os.Getenv("DEVICE_ID"))

	fallbackPlaylist = FallbackPlaylist{
		URI:    spotify.URI(os.Getenv("FALLBACK_PLAYLIST")),
		Active: false,
	}
	// Set Minimum Votes
	minimumVotes = 1

	// Start Router
	r := gin.Default()

	// go pollSpotify()

	// Set Routes
	r.GET("/login", serveLoginLink)

	r.POST("/player/:action", handlePlayer)

	r.GET("/search/:searchTerm", handleSearch)

	r.GET("/callback", handleAuth)

	r.POST("/votes/:action", handleVote)

	r.GET("/songs", getSongs)
	r.GET("/songs/:songUri", getSongByUri)
	r.POST("/songs/:action", handleSong)

	r.GET("/device/all", getAllDeviceIds)
	r.GET("/device", getCurrentDeviceId)
	r.POST("/device", setDeviceId)

	r.Run(":8888")

}
