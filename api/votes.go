package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify/v2"
)

func handleVoteSkip(c *gin.Context) {
	var playerState *spotify.PlayerState
	var track Track

	ctx := c.Request.Context()
	playerState, _ = client.PlayerState(ctx)

	if !currentDevice.Active || !playerState.Playing {
		c.JSON(http.StatusInternalServerError, "Spotify not Active")
		return
	}

	voteToSkipCurrent = voteToSkipCurrent - 1
	if voteToSkipCurrent == 0 {
		if !fallbackPlaylist.Active {
			if err := db.First(&track, Track{URI: currentTrackURI}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, "Track Not Found")
				return
			}
			if err := db.Unscoped().Delete(&Track{}, Track{URI: currentTrackURI}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		}

		voteToSkipCurrent = voteToSkipDefault
		newTrack, _ := getNextSong()
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
	c.JSON(http.StatusOK, voteToSkipCurrent)
}

func handleVote(c *gin.Context) {
	var handleVoteInput HandleVoteInput
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
		if track.Votes <= voteStandardMinimumVotes {
			if err := db.Unscoped().Delete(&Track{}, Track{URI: handleVoteInput.URI}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
			// If currently playing is voted off - play next in queue
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
		} else {
			if err := db.Model(&track).Update("votes", track.Votes).Error; err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		}
	}
	c.JSON(http.StatusOK, "Ok")
}
