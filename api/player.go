package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// var currentTrackPlayOptions spotify.PlayOptions

var pollingSpotify = false

func handlePlayer(c *gin.Context) {
	// var handlePlayerInput HandlePlayerInput
	// var playerState *spotify.PlayerState
	var err error
	// var track Track

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
		go pollSpotify(authInput)
	case "play":
		if pollingSpotify {
			err = client.PlayOpt(ctx, &playerOpt)
		} else {
			go pollSpotify(authInput)
		}
	case "pause":
		err = client.PauseOpt(ctx, &playerOpt)
	case "skip":
		track, _ := getNextSong(client, currentTrackURI)
		playerOpt.URIs = []spotify.URI{track.URI}
		err = client.NextOpt(ctx, &playerOpt)
		// Debug - do proper error handling
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, err)
		}
		if !fallbackPlaylist.Active {
			if err := db.First(&track, Track{URI: currentTrackURI}).Error; err != nil {
				c.JSON(http.StatusNotFound, "Track Not Found")
				return
			}
			if err := db.Unscoped().Delete(&track).Error; err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		}

	}
	// Debug - do proper error handling
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusAccepted, "Ok")
}

func getAllDeviceIds(c *gin.Context) {
	authHeader := c.Request.Header.Get("authorization")
	authToken := strings.Split(authHeader, " ")[1]

	authInput := oauth2.Token{
		AccessToken: authToken,
	}
	client := spotify.New(auth.Client(c.Request.Context(), &authInput))

	ctx := c.Request.Context()
	devices, err := client.PlayerDevices(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusAccepted, devices)
}

func pollSpotify(authInput oauth2.Token) {

	var track Track

	pollingSpotify = true

	client := spotify.New(auth.Client(context.Background(), &authInput))

	// DEBUG - Currently just restarts playback based on next item (will skip current song)
	//maybe dont do a stop, check if playing, set fallback and current track URI and run the loop
	nextSong, _ := getNextSong(client)
	err := client.PlayOpt(context.Background(), &spotify.PlayOptions{DeviceID: &deviceID, URIs: []spotify.URI{nextSong.URI}})
	if err != nil {
		// DEBUG - error handling
		fmt.Println("SOMETHING WENT WRONG STARTING PLAYER")
		fmt.Println(err)
	}

	fallbackPlaylist.Active = nextSong.FromFallBackPlaylist
	currentTrackURI = nextSong.URI

	// r := rand.New(rand.NewSource(99))
	c := time.Tick(10 * time.Second)

	fmt.Println("STARTING JUKEBOX WITH DEVICE: " + deviceID)
	fmt.Println("STARTING SONG: " + nextSong.Name + " - " + nextSong.Artist)

	// Start the main Loop
	for _ = range c {
		playerState, err := client.PlayerState(context.Background())
		if err != nil {
			fmt.Println("SOMETHING WENT WRONG STARTING PLAYER")
			fmt.Println(err)
		}
		fmt.Println("CURRENT SONG: " + playerState.Item.Name + " - " + playerState.Item.Artists[0].Name)
		fmt.Println("CURRENT PROGRESS: " + strconv.Itoa(playerState.Progress))
		fmt.Println("FALLBACK STATUS: " + strconv.FormatBool(fallbackPlaylist.Active))

		if playerState.Progress == 0 {
			fmt.Println("LOADING NEXT SONG")
			// Remove the track
			// DEBUG - if fallback dont run
			if !fallbackPlaylist.Active {
				if fallbackPlaylist.AddToPlaylist {
					fmt.Println("ADDING TRACK TO FALLBACK PLAYLIST: " + currentTrackURI)
					_, err := client.AddTracksToPlaylist(context.Background(), fallbackPlaylist.ID, spotify.ID(strings.Replace(string(currentTrackURI), "spotify:track:", "", -1)))
					if err != nil {
						fmt.Println("SOMETHING WENT WRONG GETTING NEXT SONG")
						fmt.Println(err)
					}
				}
				fmt.Println("REMOVING TRACK FROM QUEUE: " + currentTrackURI)
				if err := db.First(&track, Track{URI: currentTrackURI}).Error; err != nil {
					fmt.Println(err)
				}
				if err := db.Unscoped().Delete(&track).Error; err != nil {
					fmt.Println(err)
				}
			}
			// Get the next track
			track, err = getNextSong(client)
			if err != nil {
				fmt.Println("SOMETHING WENT WRONG GETTING NEXT SONG")
				fmt.Println(err)
			}
			currentTrackURI = track.URI
			fallbackPlaylist.Active = track.FromFallBackPlaylist

			if fallbackPlaylist.Active {
				fmt.Println("No More Tracks - Using fall back playlist")
			}

			fmt.Println("NEXT SONG: " + track.Name + " - " + track.Artist)

			playerOpt := spotify.PlayOptions{
				DeviceID: &deviceID,
				URIs:     []spotify.URI{track.URI},
			}
			err = client.PlayOpt(context.Background(), &playerOpt)

			if err != nil {
				fmt.Println(err)
			}
		}

	}
}

func getCurrentDeviceId(c *gin.Context) {
	c.JSON(http.StatusAccepted, deviceID)
}

func setDeviceId(c *gin.Context) {
	var setDeviceIdInput SetDeviceIdInput
	if err := c.ShouldBindJSON(&setDeviceIdInput); err != nil {
		c.JSON(http.StatusInternalServerError, "Cannot Marshal JSON")
		return
	}
	if setDeviceIdInput.DeviceId == "" {
		c.JSON(http.StatusInternalServerError, "Device ID is required.")
		return
	}
	deviceID = setDeviceIdInput.DeviceId
	c.JSON(http.StatusAccepted, deviceID)
}
