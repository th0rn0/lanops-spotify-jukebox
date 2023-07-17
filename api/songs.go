package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-sqlite3"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

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
		track, err := client.GetTrack(ctx, spotify.ID(strings.Replace(string(handleSongInput.URI), "spotify:track:", "", -1)))
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		// Get Track Images
		trackImages := []TrackImage{}
		for _, image := range track.Album.Images {
			thisImage := TrackImage{
				URL:      image.URL,
				Height:   image.Height,
				Width:    image.Width,
				TrackURI: track.URI,
			}
			trackImages = append(trackImages, thisImage)
		}
		if err := db.Create(&Track{URI: handleSongInput.URI, Name: track.Name, Artist: track.Artists[0].Name, Votes: 1, Images: trackImages}).Error; err != nil {
			if err.(sqlite3.Error).Code == 19 {
				c.JSON(http.StatusBadRequest, "Song Already Exists")
				return
			}
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		// fallbackPlaylist.Active = false
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

func getSongByUri(c *gin.Context) {
	var track Track
	if err := db.Preload("Images").First(&track, Track{URI: spotify.URI(c.Param("songUri"))}).Error; err != nil {
		// DEBUG - Correct Responses
		c.JSON(http.StatusNotFound, "Track Not Found")
		return
	}
	c.JSON(http.StatusAccepted, track)
}

func getSongs(c *gin.Context) {
	var tracks []Track
	if err := db.Preload("Images").Order("votes DESC").Find(&tracks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusAccepted, tracks)
}

// func getSongNext() {

// }

func getSongCurrent(c *gin.Context) {
	var playerState *spotify.PlayerState
	var err error
	// var track Track

	authHeader := c.Request.Header.Get("authorization")
	authToken := strings.Split(authHeader, " ")[1]

	authInput := oauth2.Token{
		AccessToken: authToken,
	}

	ctx := c.Request.Context()

	client := spotify.New(auth.Client(ctx, &authInput))

	playerState, err = client.PlayerState(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusAccepted, playerState.Item)
}
