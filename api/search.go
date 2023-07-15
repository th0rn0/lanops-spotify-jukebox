package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

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
	fmt.Print(results.Tracks.Tracks[0].Album.Images)

	searchOutput := SearchOutput{}

	// handle artist results
	if results.Artists != nil {
		fmt.Println("Artists:")
		for _, artist := range results.Artists.Artists {
			artistInfo := ArtistSearchOutput{
				Name: artist.Name,
				ID:   artist.ID.String(),
			}
			searchOutput.Artists = append(searchOutput.Artists, artistInfo)
			fmt.Println(artistInfo)
			fmt.Println("   ", artist.Name)
		}
	}

	// handle song results
	if results.Tracks != nil {
		fmt.Println("Tracks:")
		for _, track := range results.Tracks.Tracks {
			trackInfo := TrackSearchOutput{
				Name:   track.Name,
				Artist: track.Artists[0].Name,
				ID:     track.ID.String(),
				Images: track.Album.Images,
			}
			searchOutput.Tracks = append(searchOutput.Tracks, trackInfo)
			fmt.Println(trackInfo)
			fmt.Println("   ", track.Name)
		}
	}
	c.JSON(http.StatusOK, searchOutput)
}
