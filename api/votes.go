package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify/v2"
)

func handleVote(c *gin.Context) {
	// DEBUG - change this
	var handleVoteInput HandleTrackInput
	var playerState *spotify.PlayerState
	var track Track

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
		track.Votes = track.Votes - 1
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
			if err := db.Model(&track).Update("votes", track.Votes).Error; err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		}
	}
	c.JSON(http.StatusOK, "Ok")
}
