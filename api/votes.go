package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func handleVote(c *gin.Context) {
	var handleVoteInput HandleSongInput
	var playerState *spotify.PlayerState
	var track Track

	authHeader := c.Request.Header.Get("authorization")
	authToken := strings.Split(authHeader, " ")[1]

	authInput := oauth2.Token{
		AccessToken: authToken,
	}
	client := spotify.New(auth.Client(c.Request.Context(), &authInput))

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
			if err := db.Model(&track).Update("votes", track.Votes-1).Error; err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		}
	}
	c.JSON(http.StatusOK, "Ok")
}
