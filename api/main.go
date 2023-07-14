package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mattn/go-sqlite3"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"

	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify/v2"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	deviceID     spotify.ID
	db           *gorm.DB
	auth         *spotifyauth.Authenticator
	minimumVotes int64

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

	// Set Device ID
	deviceID = spotify.ID(os.Getenv("DEVICE_ID"))

	// Set Minimum Votes
	minimumVotes = 1

	// Start Router
	r := gin.Default()

	// Set Routes
	r.GET("/login", serveLoginLink)

	r.POST("/player/:action", handlePlayer)

	r.GET("/search/:searchTerm", handleSearch)

	r.GET("/callback", handleAuth)

	r.POST("/votes/:action", handleVote)

	r.GET("/songs", getSongs)
	r.GET("/songs/:songUri", getSongByUri)
	r.POST("/songs/:action", handleSong)

	r.GET("/device/all", getDeviceIds)

	r.Run(":8888")
}

func handlePlayer(c *gin.Context) {
	var handlePlayerInput HandlePlayerInput
	var playerState *spotify.PlayerState
	var err error
	var track Track

	authHeader := c.Request.Header.Get("authorization")
	authToken := strings.Split(authHeader, " ")[1]

	authInput := oauth2.Token{
		AccessToken: authToken,
	}
	client := spotify.New(auth.Client(c.Request.Context(), &authInput))

	ctx := c.Request.Context()
	action := c.Param("action")

	// DEBUG
	fmt.Println("Got request for:", action)

	playerOpt := spotify.PlayOptions{
		DeviceID: &deviceID,
	}

	switch action {
	case "start":
		playerState, _ = client.PlayerState(ctx)
		if playerState.Playing {
			c.JSON(http.StatusBadRequest, "Player Already Started")
			return
		}
		// DEBUG - Handle Error
		track, err = getNextSongByVotes()
		fmt.Println(track)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		playerOpt = spotify.PlayOptions{
			DeviceID: &deviceID,
			URIs:     []spotify.URI{track.URI},
		}
		err = client.PlayOpt(ctx, &playerOpt)
	case "play":
		err = client.PlayOpt(ctx, &playerOpt)
	case "pause":
		err = client.PauseOpt(ctx, &playerOpt)
	case "next":
		err = client.NextOpt(ctx, &playerOpt)
	case "previous":
		err = client.PreviousOpt(ctx, &playerOpt)
	case "shuffle":
		playerState.ShuffleState = !playerState.ShuffleState
		err = client.ShuffleOpt(ctx, playerState.ShuffleState, &playerOpt)
	case "song":
		if err := c.ShouldBindJSON(&handlePlayerInput); err != nil {
			c.JSON(http.StatusInternalServerError, "Cannot Marshal JSON")
			return
		}
		if handlePlayerInput.URI == "" {
			c.JSON(http.StatusInternalServerError, "URI is required.")
			return
		}
		playerOpt = spotify.PlayOptions{
			DeviceID: &deviceID,
			URIs:     []spotify.URI{handlePlayerInput.URI},
		}
		err = client.PlayOpt(ctx, &playerOpt)
	}
	// Debug - do proper error handling
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusAccepted, "Ok")
}

func serveLoginLink(c *gin.Context) {
	url := auth.AuthURL(state)
	c.JSON(http.StatusOK, url)
}

func handleAuth(c *gin.Context) {
	tok, err := auth.Token(c.Request.Context(), state, c.Request)
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, err)
	}
	if st := c.Request.FormValue("state"); st != state {
		log.Fatalf("State mismatch: %s != %s\n", st, state)
		c.JSON(http.StatusNotFound, err)
	}
	// Return Auth to client
	c.JSON(http.StatusOK, LoginToken{
		AccessToken:  tok.AccessToken,
		TokenType:    tok.TokenType,
		RefreshToken: tok.RefreshToken,
		Expiry:       tok.Expiry.String(),
	})
}

func handleSearch(c *gin.Context) {
	authHeader := c.Request.Header.Get("authorization")
	authToken := strings.Split(authHeader, " ")[1]

	authInput := oauth2.Token{
		AccessToken: authToken,
	}
	client := spotify.New(auth.Client(c.Request.Context(), &authInput))

	ctx := c.Request.Context()
	searchTerm := c.Param("searchTerm")
	fmt.Println(searchTerm)

	// results, err := client.Search(ctx, searchTerm, spotify.SearchTypeArtist|spotify.SearchTypeTrack)
	results, err := client.Search(ctx, searchTerm, spotify.SearchTypeTrack)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(results)

	searchOutput := SearchOutput{}

	// handle artist results
	if results.Artists != nil {
		fmt.Println("Artists:")
		for _, item := range results.Artists.Artists {
			artistInfo := ArtistSearchOutput{
				Name: item.Name,
				ID:   item.ID.String(),
			}
			searchOutput.Artists = append(searchOutput.Artists, artistInfo)
			fmt.Println(artistInfo)
			fmt.Println("   ", item.Name)
		}
	}

	// handle song results
	if results.Tracks != nil {
		fmt.Println("Tracks:")
		for _, item := range results.Tracks.Tracks {
			trackInfo := TrackSearchOutput{
				Name:   item.Name,
				Artist: item.Artists[0].Name,
				ID:     item.ID.String(),
			}
			searchOutput.Tracks = append(searchOutput.Tracks, trackInfo)
			fmt.Println(trackInfo)
			fmt.Println("   ", item.Name)
		}
	}
	c.JSON(http.StatusOK, searchOutput)
}

func handleSong(c *gin.Context) {
	authHeader := c.Request.Header.Get("authorization")
	authToken := strings.Split(authHeader, " ")[1]

	authInput := oauth2.Token{
		AccessToken: authToken,
	}
	client := spotify.New(auth.Client(c.Request.Context(), &authInput))

	var handleSongInput HandleSongInput
	var playerState *spotify.PlayerState
	var returnStatus = http.StatusCreated
	var track Track

	ctx := c.Request.Context()
	action := c.Param("action")

	if err := c.ShouldBindJSON(&handleSongInput); err != nil {
		c.JSON(http.StatusInternalServerError, "Cannot Marshal JSON")
		return
	}
	if handleSongInput.URI == "" {
		c.JSON(http.StatusInternalServerError, "URI is required.")
		return
	}

	switch action {
	case "add":
		// Get Track Info
		trackId := strings.Replace(string(handleSongInput.URI), "spotify:track:", "", -1)
		track, err := client.GetTrack(ctx, spotify.ID(trackId))
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		if err := db.Create(&Track{URI: handleSongInput.URI, Name: track.Name, Artist: track.Artists[0].Name, Votes: 1}).Error; err != nil {
			if err.(sqlite3.Error).Code == 19 {
				c.JSON(http.StatusBadRequest, "Song Already Exists")
				return
			}
			c.JSON(http.StatusInternalServerError, err)
			return
		}
	case "remove":
		playerState, _ = client.PlayerState(ctx)
		if err := db.First(&track, Track{URI: handleSongInput.URI}).Error; err != nil {
			c.JSON(http.StatusNotFound, "Track Not Found")
			return
		}
		if err := db.Unscoped().Delete(&track).Error; err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		// If currently playing is removed - play next in queue
		if playerState.Playing && playerState.Item.URI == track.URI {
			newTrack, _ := getNextSongByVotes()
			playerOpt := spotify.PlayOptions{
				DeviceID: &deviceID,
				URIs:     []spotify.URI{newTrack.URI},
			}
			err := client.PlayOpt(ctx, &playerOpt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		}
		returnStatus = http.StatusAccepted
	}
	// DEBUG - make this the proper response http.StatusCreated
	c.JSON(returnStatus, "Ok")
}

func handleVote(c *gin.Context) {
	var handleVoteInput HandleSongInput
	var playerState *spotify.PlayerState
	var track Track

	authHeader := c.Request.Header.Get("authorization")
	authToken := strings.Split(authHeader, " ")[1]

	authInput := oauth2.Token{
		AccessToken: authToken,
	}
	client := spotify.New(auth.Client(c.Request.Context(), &authInput))

	ctx := c.Request.Context()
	action := c.Param("action")

	if err := c.ShouldBindJSON(&handleVoteInput); err != nil {
		c.JSON(http.StatusInternalServerError, "Cannot Marshal JSON")
		return
	}
	if handleVoteInput.URI == "" {
		c.JSON(http.StatusInternalServerError, "URI is required.")
		return
	}

	if err := db.First(&track, Track{URI: handleVoteInput.URI}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, "Track Not Found")
		return
	}

	switch action {
	case "add":
		if err := db.Model(&track).Update("votes", track.Votes+1).Error; err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
	case "remove":
		playerState, _ = client.PlayerState(ctx)
		if track.Votes <= minimumVotes {
			if err := db.Unscoped().Delete(&Track{}, Track{URI: handleVoteInput.URI}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
			// If currently playing is voted off - play next in queue
			if playerState.Playing && playerState.Item.URI == track.URI {
				newTrack, _ := getNextSongByVotes()
				playerOpt := spotify.PlayOptions{
					DeviceID: &deviceID,
					URIs:     []spotify.URI{newTrack.URI},
				}
				err := client.PlayOpt(ctx, &playerOpt)
				if err != nil {
					c.JSON(http.StatusInternalServerError, err)
					return
				}
			}
		} else {
			if err := db.Model(&track).Update("votes", track.Votes-1).Error; err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		}
	}
	c.JSON(http.StatusOK, "Ok")
}

func getDeviceIds(c *gin.Context) {
	authHeader := c.Request.Header.Get("authorization")
	authToken := strings.Split(authHeader, " ")[1]

	authInput := oauth2.Token{
		AccessToken: authToken,
	}
	client := spotify.New(auth.Client(c.Request.Context(), &authInput))

	ctx := c.Request.Context()
	devices, err := client.PlayerDevices(ctx)
	// client.Get
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusAccepted, devices)
}

// Getters
func getSongByUri(c *gin.Context) {
	var track Track
	if err := db.First(&track, Track{URI: spotify.URI(c.Param("songUri"))}).Error; err != nil {
		// DEBUG - Correct Responses
		c.JSON(http.StatusNotFound, "Track Not Found")
		return
	}
	c.JSON(http.StatusAccepted, track)
}

func getSongs(c *gin.Context) {
	var tracks []Track
	// DEBUG - ORDER BY NEEDED
	if err := db.Find(&tracks).Order("votes ASC").Error; err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusAccepted, tracks)
}

// func getSongNext() {

// }

// func getSongCurrent() {

// }

// HELPERS

func getNextSongByVotes() (Track, error) {
	var track Track

	// if err := db.Raw("SELECT MAX(votes) FROM tracks").First(&track).Error; err != nil {
	if err := db.Raw("SELECT * FROM tracks WHERE votes = ( SELECT MAX(votes) FROM tracks )").First(&track).Error; err != nil {
		return track, err
	}
	return track, nil
}
