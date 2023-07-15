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

func handlePlayer(c *gin.Context) {
	var handlePlayerInput HandlePlayerInput
	var playerState *spotify.PlayerState
	var err error
	var track Track

	authHeader := c.Request.Header.Get("authorization")
	authToken := strings.Split(authHeader, " ")[1]

	authInput := oauth2.Token{
		AccessToken: authToken,
	}
	client := spotify.New(auth.Client(c.Request.Context(), &authInput))

	ctx := c.Request.Context()
	action := c.Param("action")

	// DEBUG
	fmt.Println("Got request for:", action)

	playerOpt := spotify.PlayOptions{
		DeviceID: &deviceID,
	}

	switch action {
	case "start":
		playerState, _ = client.PlayerState(ctx)
		if playerState.Playing {
			c.JSON(http.StatusBadRequest, "Player Already Started")
			return
		}
		// DEBUG - Handle Error
		track, err = getNextSongByVotes()
		fmt.Println(track)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		playerOpt = spotify.PlayOptions{
			DeviceID: &deviceID,
			URIs:     []spotify.URI{track.URI},
		}
		err = client.PlayOpt(ctx, &playerOpt)
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
		if err := c.ShouldBindJSON(&handlePlayerInput); err != nil {
			c.JSON(http.StatusInternalServerError, "Cannot Marshal JSON")
			return
		}
		if handlePlayerInput.URI == "" {
			c.JSON(http.StatusInternalServerError, "URI is required.")
			return
		}
		playerOpt = spotify.PlayOptions{
			DeviceID: &deviceID,
			URIs:     []spotify.URI{handlePlayerInput.URI},
		}
		err = client.PlayOpt(ctx, &playerOpt)
	}
	// Debug - do proper error handling
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusAccepted, "Ok")
}

func getDeviceIds(c *gin.Context) {
	authHeader := c.Request.Header.Get("authorization")
	authToken := strings.Split(authHeader, " ")[1]

	authInput := oauth2.Token{
		AccessToken: authToken,
	}
	client := spotify.New(auth.Client(c.Request.Context(), &authInput))

	ctx := c.Request.Context()
	devices, err := client.PlayerDevices(ctx)
	// client.Get
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusAccepted, devices)
}
