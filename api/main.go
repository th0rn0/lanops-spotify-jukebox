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
	deviceID spotify.ID
	db       *gorm.DB
	auth     *spotifyauth.Authenticator

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

	// Start Router
	r := gin.Default()

	// Set Routes
	r.GET("/login", serveLoginLink)
	r.POST("/player/:action", handlePlayer)
	r.GET("/search/:searchTerm", handleSearch)
	r.GET("/callback", handleAuth)
	r.GET("/votes", getVotes)

	r.GET("/songs", getSongs)
	r.POST("/songs/:action", handleSong)

	r.Run(":8888")
}

func handlePlayer(c *gin.Context) {
	var playerSongInput PlayerSongInput
	var playerState *spotify.PlayerState
	var err error

	authHeader := c.Request.Header.Get("authorization")
	authToken := strings.Split(authHeader, " ")[1]

	authInput := oauth2.Token{
		AccessToken: authToken,
	}
	client := spotify.New(auth.Client(c.Request.Context(), &authInput))

	ctx := c.Request.Context()
	action := c.Param("action")
	fmt.Println("Got request for:", action)

	playerOpt := spotify.PlayOptions{
		DeviceID: &deviceID,
	}

	switch action {
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
		if err := c.ShouldBindJSON(&playerSongInput); err != nil {
			c.JSON(http.StatusInternalServerError, "Cannot Marshal JSON")
			return
		}
		if playerSongInput.URI == "" {
			c.JSON(http.StatusInternalServerError, "URI is required.")
			return
		}
		playerOpt = spotify.PlayOptions{
			DeviceID: &deviceID,
			URIs:     []spotify.URI{playerSongInput.URI},
		}
		err = client.PlayOpt(ctx, &playerOpt)
	}
	if err != nil {
		log.Print(err)
	}

	// c.Writer.Header().Set("Content-Type", "text/html")
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

	var handleSongInput HandleSongInput
	var result *gorm.DB
	// var err error

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
		client := spotify.New(auth.Client(c.Request.Context(), &authInput))
		trackId := strings.Replace(string(handleSongInput.URI), "spotify:track:", "", -1)
		track, err := client.GetTrack(ctx, spotify.ID(trackId))
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		result = db.Create(&Track{URI: handleSongInput.URI, Name: track.Name, Artist: track.Artists[0].Name, Votes: 1})
	case "remove":
		fmt.Println("we are here")
		track := db.First(&Track{}, Track{URI: handleSongInput.URI})
		// DEBUG - bit hacky - maybe do some better error handling instead of assuming no record found
		if track.Error != nil {
			c.JSON(http.StatusNotFound, "Track Not Found")
			return
		}
		result = db.Unscoped().Delete(&Track{}, Track{URI: handleSongInput.URI})
	}
	if result.Error != nil {
		if result.Error.(sqlite3.Error).Code == 19 {
			c.JSON(http.StatusBadRequest, "Song Already Exists")
			return
		}

		c.JSON(http.StatusInternalServerError, "Something Went Wrong. Contact an Adult")
	}
	c.JSON(http.StatusAccepted, "Ok")
}

// Getters
func getVotes(c *gin.Context) {
	// fmt.Println("asdasdasd")
	// var songQueue Song
	// result := db.Find(&songQueue) // find product with integer primary key
	// fmt.Println(&result.RowsAffected)
}

func getSongs(c *gin.Context) {
	var tracks []Track
	result := db.Find(&tracks)
	c.JSON(http.StatusAccepted, result)
}
