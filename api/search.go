package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify/v2"
)

func handleSearch(c *gin.Context) {
	var searchOutput SearchOutput

	ctx := c.Request.Context()
	searchTerm := c.Param("searchTerm")

	results, err := client.Search(ctx, searchTerm, spotify.SearchTypeTrack)
	if err != nil {
		logger.Fatal().Err(err)
	}

	// handle artist results
	if results.Artists != nil {
		for _, artist := range results.Artists.Artists {
			artistInfo := ArtistSearchOutput{
				Name: artist.Name,
				ID:   artist.ID.String(),
			}
			searchOutput.Artists = append(searchOutput.Artists, artistInfo)
		}
	}

	// handle track results
	if results.Tracks != nil {
		for _, track := range results.Tracks.Tracks {
			trackInfo := TrackSearchOutput{
				Name:   track.Name,
				Artist: track.Artists[0].Name,
				ID:     track.ID.String(),
				Images: track.Album.Images,
			}
			searchOutput.Tracks = append(searchOutput.Tracks, trackInfo)
		}
	}
	c.JSON(http.StatusOK, searchOutput)
}
