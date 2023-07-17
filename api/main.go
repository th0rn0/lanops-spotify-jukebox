package main

import (
	"log"
	"os"
	"strconv"
	"strings"

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
	// DEBUG - make this currentTrack and pull in all info
	currentTrackURI spotify.URI
	client          *spotify.Client
)

var (
	state          = "spotifyJukeBox"
	pollingSpotify = false
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
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadCurrentlyPlaying,
			spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserModifyPlaybackState,
			spotifyauth.ScopePlaylistModifyPrivate,
			spotifyauth.ScopePlaylistModifyPublic,
		),
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

	addToPlaylist, _ := strconv.ParseBool(os.Getenv("FALLBACK_PLAYLIST_ADD_QUEUED"))
	fallbackPlaylist = FallbackPlaylist{
		URI:           spotify.URI(os.Getenv("FALLBACK_PLAYLIST_URI")),
		ID:            spotify.ID(strings.Replace(os.Getenv("FALLBACK_PLAYLIST_URI"), "spotify:playlist:", "", -1)),
		Active:        false,
		AddToPlaylist: addToPlaylist,
	}

	// Set Minimum Votes
	minimumVotes, _ = strconv.ParseInt(os.Getenv("MINIMUM_VOTES"), 10, 64)

	// Start Router
	r := gin.Default()

	// DEBUG - do this here?
	// go pollSpotify()

	// Set Routes
	r.GET("/login", serveLoginLink)

	r.POST("/player/:action", handlePlayer)

	r.GET("/search/:searchTerm", handleSearch)

	r.GET("/callback", handleAuth)

	r.POST("/votes/:action", handleVote)

	r.GET("/songs", getSongs)
	// DEBUG - CHANGE ME
	r.GET("/songs/current", getSongCurrent)
	r.GET("/songs/:songUri", getSongByUri)
	r.POST("/songs/:action", handleSong)

	r.GET("/device/all", getAllDeviceIds)
	r.GET("/device", getCurrentDeviceId)
	r.POST("/device", setDeviceId)
	// go func() {
	// 	r.Run(":8888")
	// }()
	r.Run(":8888")
}
