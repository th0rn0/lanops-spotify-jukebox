package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"

	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify/v2"
)

var redirectURI = os.Getenv("CALLBACK_URL")
var deviceID = spotify.ID(os.Getenv("DEVICE_ID"))

var (
	auth = spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadCurrentlyPlaying, spotifyauth.ScopeUserReadPlaybackState, spotifyauth.ScopeUserModifyPlaybackState),
	)
	// ch    = make(chan *spotify.Client)
	state = "abc123"
)

func main() {
	r := gin.Default()

	// Routes
	r.GET("/login", serveLoginLink)
	r.POST("/player/:action", handlePlayer)
	r.GET("/search/:searchTerm", handleSearch)
	r.GET("/callback", completeAuth)

	r.Run(":8888")
}

func handlePlayer(c *gin.Context) {
	var playerSongInput PlayerSongInput
	var playerState *spotify.PlayerState

	if err := c.ShouldBindJSON(&playerSongInput); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

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

	log.Print(playerOpt)

	var err error
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

	searchResults := []TrackResult{}

	// // handle artist results
	// if results.Artists != nil {
	// 	fmt.Println("Artists:")
	// 	for _, item := range results.Albums.Albums {
	// 		artistInfo := artistResult{
	// 			Name: item.Name,
	// 			ID:   item.ID.String(),
	// 		}

	// 		searchResults = append(searchResults, artistInfo)
	// 		fmt.Println(artistInfo)
	// 		fmt.Println("   ", item.Name)
	// 	}
	// }

	// handle song results
	if results.Tracks != nil {
		fmt.Println("Tracks:")
		for _, item := range results.Tracks.Tracks {
			trackInfo := TrackResult{
				Name:   item.Name,
				Artist: item.Artists[0].Name,
				ID:     item.ID.String(),
			}

			searchResults = append(searchResults, trackInfo)
			fmt.Println(trackInfo)
			fmt.Println("   ", item.Name)
		}
	}
	c.JSON(http.StatusOK, searchResults)
}
