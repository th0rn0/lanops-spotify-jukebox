package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"

	"github.com/zmb3/spotify/v2"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var (
	currentDevice            spotify.PlayerDevice
	fallbackPlaylist         FallbackPlaylist
	db                       *gorm.DB
	auth                     *spotifyauth.Authenticator
	currentTrackURI          spotify.URI
	client                   *spotify.Client
	oauthToken               LoginToken
	logger                   zerolog.Logger
	adminPassword            string
	voteStandardRateLimit    uint64
	voteStandardMinimumVotes int64
	voteToSkipDefault        int64
	voteToSkipCurrent        int64
	voteToSkipEnabled        bool
)

var (
	state          = "spotifyJukeBox"
	pollingSpotify = false
)

func init() {
	var err error

	logger = zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
	).Level(zerolog.TraceLevel).With().Timestamp().Caller().Logger()
	logger.Info().Msg("Initializing Jukebox API")

	// Env Variables
	logger.Info().Msg("Loading Environment Variables")
	godotenv.Load()

	// Load Database & Migrate the schema
	db, err = gorm.Open(sqlite.Open(os.Getenv("DB_PATH")), &gorm.Config{})
	logger.Info().Msg("Connecting to Database")
	if err != nil {
		logger.Fatal().Err(err).Msg("Error Connecting to Database")
	}
	db.AutoMigrate(&Track{})
	db.AutoMigrate(&TrackImage{})
	db.AutoMigrate(&Device{})
	db.AutoMigrate(&LoginToken{})
	db.AutoMigrate(&BannedWord{})
	db.AutoMigrate(&BannedTrack{})

	// Load Banned Words
	bannedWordsFile, err := os.Open("templates/banned-words.txt")
	if err != nil {
		panic(err)
	}
	defer bannedWordsFile.Close()

	scannerBannedWords := bufio.NewScanner(bannedWordsFile)
	for scannerBannedWords.Scan() {
		if !containsBannedWord(scannerBannedWords.Text()) {
			_, err := addBannedWord(fmt.Sprintf("%v", scannerBannedWords.Text()))
			if err != nil {
				logger.Fatal().Err(err).Msg("Cannot Set Banned Words")
			}
		}
	}

	if err := scannerBannedWords.Err(); err != nil {
		log.Fatal(err)
	}

	// Load Banned Tracks
	bannedTracksFile, err := os.Open("templates/banned-tracks.txt")
	if err != nil {
		panic(err)
	}
	defer bannedTracksFile.Close()

	scannerBannedTracks := bufio.NewScanner(bannedTracksFile)
	for scannerBannedTracks.Scan() {
		if !isBannedTrack(spotify.URI(fmt.Sprintf("%v", scannerBannedTracks.Text()))) {
			_, err := addBannedTrack(spotify.URI(fmt.Sprintf("%v", scannerBannedTracks.Text())))
			if err != nil {
				logger.Fatal().Err(err).Msg("Cannot Set Banned Words")
			}
		}
	}

	if err := scannerBannedTracks.Err(); err != nil {
		log.Fatal(err)
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

	// Set Spotify Login Token
	dbLoginToken := LoginToken{}
	if err := db.First(&dbLoginToken).Error; err != nil {
		// Assume no Login is Set
		logger.Info().Msg("NO LOGIN SET")
	} else {
		oauthToken.AccessToken = dbLoginToken.AccessToken
		oauthToken.TokenType = dbLoginToken.TokenType
		oauthToken.RefreshToken = dbLoginToken.RefreshToken
		oauthToken.Expiry = dbLoginToken.Expiry
		client = spotify.New(auth.Client(context.TODO(), &oauth2.Token{
			AccessToken:  dbLoginToken.AccessToken,
			TokenType:    dbLoginToken.TokenType,
			RefreshToken: dbLoginToken.RefreshToken,
			Expiry:       dbLoginToken.Expiry,
		}))
		logger.Info().Msg("LOGIN SET")
	}

	// Set Device ID
	dbDevice := Device{}
	if err := db.First(&dbDevice).Error; err != nil {
		// Assume no Device is Set
		logger.Info().Msg("NO DEVICE SET")
	} else {
		currentDevice.ID = dbDevice.ID
		currentDevice.Active = false
		currentDevice.Name = dbDevice.Name
		currentDevice.Type = dbDevice.Type
		logger.Info().Msg("DEVICE SET")
		logger.Info().Msg(dbDevice.Name)
	}

	// Set Minimum Votes
	voteStandardMinimumVotes, _ = strconv.ParseInt(os.Getenv("VOTE_STANDARD_MINIMUM_TO_REMOVE"), 10, 64)

	voteToSkipDefault, _ = strconv.ParseInt(os.Getenv("VOTES_TO_SKIP"), 10, 64)
	voteToSkipCurrent = voteToSkipDefault

	// Set Vote Method
	voteToSkipEnabled = false
	if os.Getenv("VOTE_METHOD") == "skip" {
		voteToSkipEnabled = true
	}

	// Set Fallback Playlist
	addToPlaylist, _ := strconv.ParseBool(os.Getenv("FALLBACK_PLAYLIST_ADD_QUEUED"))
	fallbackPlaylist = FallbackPlaylist{
		URI:           spotify.URI(os.Getenv("FALLBACK_PLAYLIST_URI")),
		ID:            spotify.ID(strings.Replace(os.Getenv("FALLBACK_PLAYLIST_URI"), "spotify:playlist:", "", -1)),
		Active:        false,
		AddToPlaylist: addToPlaylist,
	}

	// Set Rate Limiting
	voteStandardRateLimit, _ = strconv.ParseUint(os.Getenv("VOTE_STANDARD_MAXIMUM_PER_HOUR"), 10, 32)

	// Admin Password
	adminPassword = os.Getenv("ADMIN_PASSWORD")

	logger.Info().Msg("Initalization Complete")
}

func main() {
	logger.Info().Msg("Starting Jukebox API")

	// Start Listeners and Polling
	logger.Info().Msg("Starting GIN Web Server")
	// Set Rate Limiting
	rateLimitVoteStandardMiddleWareStore := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Hour,
		Limit: uint(voteStandardRateLimit),
	})

	rateLimitVoteStandardMiddleWare := ratelimit.RateLimiter(rateLimitVoteStandardMiddleWareStore, &ratelimit.Options{
		ErrorHandler: func(c *gin.Context, info ratelimit.Info) {
			c.JSON(429, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
		},
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP() + c.Request.UserAgent()
		},
	})

	rateLimitVoteSkipMiddleWareStore := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Minute * 4,
		Limit: 1,
	})

	rateLimitVoteSkipMiddleWare := ratelimit.RateLimiter(rateLimitVoteSkipMiddleWareStore, &ratelimit.Options{
		ErrorHandler: func(c *gin.Context, info ratelimit.Info) {
			c.JSON(429, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
		},
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP() + c.Request.UserAgent()
		},
	})

	r := gin.Default()

	r.Use(cors.Default())

	authorized := r.Group("", gin.BasicAuth(gin.Accounts{
		"admin": adminPassword,
	}))

	// Attempt to reboot previous session
	go startPollingSpotifyIfActive()

	// Set Routes
	r.GET("/search/:searchTerm", handleSearch)

	if voteToSkipEnabled {
		r.POST("/votes/skip", rateLimitVoteSkipMiddleWare, handleVoteSkip)
	} else {
		r.POST("/votes/:action", rateLimitVoteStandardMiddleWare, handleVote)
	}

	r.GET("/auth/callback", handleAuth)
	r.GET("/auth/login", serveLoginLink)

	authorized.POST("/player/:action", handlePlayer)

	r.GET("/tracks", getTracks)
	r.GET("/tracks/current", getTrackCurrent)
	r.GET("/tracks/:trackUri", getTrackByUri)
	r.POST("/tracks/:action", handleTrack)
	authorized.POST("/tracks/remove", removeTrack)

	authorized.GET("/device/all", getAllDevices)
	authorized.GET("/device", getCurrentDevice)
	authorized.POST("/device", setDevice)

	r.Run(":8888")
}

func startPollingSpotifyIfActive() {
	time.Sleep(8 * time.Second)
	dbDevice := Device{}
	if err := db.First(&dbDevice).Error; err != nil {
		logger.Info().Msg("No Device Set - Please Manually Set the Device!")
	} else if dbDevice.Active {
		logger.Info().Msg("Attempting to Auto Start Spotify")
		go pollSpotify()
	}
}
