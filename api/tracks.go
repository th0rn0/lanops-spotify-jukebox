package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-sqlite3"
	"github.com/zmb3/spotify/v2"
)

func handleTrack(c *gin.Context) {
	var handleTrackInput HandleTrackInput
	// var playerState *spotify.PlayerState
	// var track Track

	ctx := c.Request.Context()
	action := c.Param("action")

	if err := c.ShouldBindJSON(&handleTrackInput); err != nil {
		c.JSON(http.StatusInternalServerError, "Cannot Marshal JSON")
		return
	}
	if handleTrackInput.URI == "" {
		c.JSON(http.StatusInternalServerError, "URI is required.")
		return
	}

	switch action {
	case "add":
		// Get Track Info
		track, err := client.GetTrack(ctx, spotify.ID(strings.Replace(string(handleTrackInput.URI), "spotify:track:", "", -1)))
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
		if err := db.Create(&Track{URI: handleTrackInput.URI, Name: track.Name, Artist: track.Artists[0].Name, Votes: 5, Images: trackImages}).Error; err != nil {
			if err.(sqlite3.Error).Code == 19 {
				c.JSON(http.StatusBadRequest, "Song Already Exists")
				return
			}
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusCreated, track)
		return
	// case "remove":
	// 	playerState, _ = client.PlayerState(ctx)
	// 	if err := db.First(&track, Track{URI: handleTrackInput.URI}).Error; err != nil {
	// 		c.JSON(http.StatusNotFound, "Track Not Found")
	// 		return
	// 	}
	// 	if err := db.Unscoped().Delete(&track).Error; err != nil {
	// 		c.JSON(http.StatusInternalServerError, err)
	// 		return
	// 	}
	// 	// If currently playing is removed - play next in queue
	// 	if playerState.Playing && playerState.Item.URI == track.URI {
	// 		newTrack, _ := getNextSongByVotes()
	// 		playerOpt := spotify.PlayOptions{
	// 			DeviceID: &currentDevice.ID,
	// 			URIs:     []spotify.URI{newTrack.URI},
	// 		}
	// 		err := client.PlayOpt(ctx, &playerOpt)
	// 		if err != nil {
	// 			c.JSON(http.StatusInternalServerError, err)
	// 			return
	// 		}
	// 	}
	// 	c.JSON(http.StatusAccepted, track)
	// 	return
	default:
		c.JSON(http.StatusBadRequest, "Unknown Action")
		return
	}
}

func removeTrack(c *gin.Context) {
	var handleTrackInput HandleTrackInput
	var playerState *spotify.PlayerState
	var track Track

	ctx := c.Request.Context()

	if err := c.ShouldBindJSON(&handleTrackInput); err != nil {
		c.JSON(http.StatusInternalServerError, "Cannot Marshal JSON")
		return
	}
	if handleTrackInput.URI == "" {
		c.JSON(http.StatusInternalServerError, "URI is required.")
		return
	}

	playerState, _ = client.PlayerState(ctx)
	if err := db.First(&track, Track{URI: handleTrackInput.URI}).Error; err != nil {
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
			DeviceID: &currentDevice.ID,
			URIs:     []spotify.URI{newTrack.URI},
		}
		err := client.PlayOpt(ctx, &playerOpt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
	}
	c.JSON(http.StatusAccepted, track)
}

func getTrackByUri(c *gin.Context) {
	var track Track

	if err := db.Preload("Images").First(&track, Track{URI: spotify.URI(c.Param("trackUri"))}).Error; err != nil {
		c.JSON(http.StatusNotFound, "Track Not Found")
		return
	}
	c.JSON(http.StatusAccepted, track)
}

func getTracks(c *gin.Context) {
	var tracks []Track

	if err := db.Preload("Images").Order("votes DESC").Find(&tracks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusAccepted, tracks)
}

func getTrackCurrent(c *gin.Context) {
	var playerState *spotify.PlayerState
	var err error

	playerState, err = client.PlayerState(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusAccepted, playerState.Item)
}
