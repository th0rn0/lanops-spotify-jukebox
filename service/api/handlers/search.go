package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify/v2"
)

type SearchOutput struct {
	Artists []ArtistSearchOutput `json:"artists"`
	Tracks  []TrackSearchOutput  `json:"tracks"`
}

type ArtistSearchOutput struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type TrackSearchOutput struct {
	Name   string          `json:"name"`
	Artist string          `json:"artist"`
	Id     string          `json:"id"`
	Images []spotify.Image `json:"images"`
}

func (s Client) Search(c *gin.Context) {
	var searchOutput SearchOutput

	ctx := c.Request.Context()
	searchTerm := c.Param("searchTerm")

	results, err := s.jbc.Search(ctx, searchTerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	// We only want Track & Artist Results
	if results.Tracks != nil {
		for _, track := range results.Tracks.Tracks {
			trackInfo := TrackSearchOutput{
				Name:   track.Name,
				Artist: track.Artists[0].Name,
				Id:     track.ID.String(),
				Images: track.Album.Images,
			}
			searchOutput.Tracks = append(searchOutput.Tracks, trackInfo)
		}
	}
	if results.Artists != nil {
		for _, artist := range results.Artists.Artists {
			artistInfo := ArtistSearchOutput{
				Name: artist.Name,
				Id:   artist.ID.String(),
			}
			searchOutput.Artists = append(searchOutput.Artists, artistInfo)
		}
	}
	c.JSON(http.StatusOK, searchOutput)
}
