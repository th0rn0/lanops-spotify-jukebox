package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"

	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify/v2"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	deviceID    spotify.ID
	db          *gorm.DB
	redirectURI = "http://localhost:8888/callback"
	auth        = spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI), // DEBUG - pull in via env
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadCurrentlyPlaying, spotifyauth.ScopeUserReadPlaybackState, spotifyauth.ScopeUserModifyPlaybackState),
	)
	// ch    = make(chan *spotify.Client)
	state = "spotifyJukeBox"
)

func main() {
	// Load Env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Load Database
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// Migrate the schema
	db.AutoMigrate(&Song{})

	// Set Device ID
	deviceID = spotify.ID(os.Getenv("DEVICE_ID"))

	// Start Router
	r := gin.Default()

	// Set Routes
	r.GET("/login", serveLoginLink)
	r.POST("/player/:action", handlePlayer)
	r.GET("/search/:searchTerm", handleSearch)
	r.GET("/callback", completeAuth)
	r.GET("/votes", getVotes)
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

	c.Writer.Header().Set("Content-Type", "text/html")
	c.JSON(http.StatusAccepted, "Ok")
}

func serveLoginLink(c *gin.Context) {
	url := auth.AuthURL(state)
	c.JSON(http.StatusOK, url)
}

func completeAuth(c *gin.Context) {
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

	results, err := client.Search(ctx, searchTerm, spotify.SearchTypeArtist|spotify.SearchTypeTrack)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(results)

	searchResults := SearchResult{}

	// handle artist results
	if results.Artists != nil {
		fmt.Println("Artists:")
		for _, item := range results.Artists.Artists {
			artistInfo := ArtistResult{
				Name: item.Name,
				ID:   item.ID.String(),
			}
			searchResults.ArtistResults = append(searchResults.ArtistResults, artistInfo)
			fmt.Println(artistInfo)
			fmt.Println("   ", item.Name)
		}
	}

	// handle song results
	if results.Tracks != nil {
		fmt.Println("Tracks:")
		for _, item := range results.Tracks.Tracks {
			trackInfo := TrackResult{
				Name:   item.Name,
				Artist: item.Artists[0].Name,
				ID:     item.ID.String(),
			}
			searchResults.TrackResults = append(searchResults.TrackResults, trackInfo)
			fmt.Println(trackInfo)
			fmt.Println("   ", item.Name)
		}
	}
	c.JSON(http.StatusOK, searchResults)
}

func getVotes(c *gin.Context) {
	// fmt.Println("asdasdasd")
	// var songQueue Song
	// result := db.Find(&songQueue) // find product with integer primary key
	// fmt.Println(&result.RowsAffected)
}

func handleSong(c *gin.Context) {
	var handleSongInput HandleSongInput
	var result *gorm.DB
	var err error

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
		result = db.Create(&Song{URI: handleSongInput.URI, Name: "asdasd", Votes: 1})
	case "remove":
		result = db.Where("uri LIKE  ?", "%"+handleSongInput.URI+"%").Delete(&Song{})
	}
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, result.Error)
		log.Print(err)
	}

	c.Writer.Header().Set("Content-Type", "text/html")
	c.JSON(http.StatusAccepted, "Ok")
}
