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

func handlePlayer(c *gin.Context) {
	var err error
	var handleTrackVolumeInput HandlePlayerVolumeInput

	ctx := c.Request.Context()
	action := c.Param("action")

	fmt.Println("Got request for:", action)

	if currentDevice.ID == "" {
		c.JSON(http.StatusInternalServerError, "No Device Set")
		return
	}

	playerOpt := spotify.PlayOptions{
		DeviceID: &currentDevice.ID,
	}

	switch action {
	case "start":
		go pollSpotify()
	case "play":
		if pollingSpotify {
			err = client.PlayOpt(ctx, &playerOpt)
			if err != nil {
				log.Print(err)
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		} else {
			go pollSpotify()
		}
	case "pause":
		err = client.PauseOpt(ctx, &playerOpt)
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
	case "vol":
		if err := c.ShouldBindJSON(&handleTrackVolumeInput); err != nil {
			c.JSON(http.StatusInternalServerError, "Cannot Marshal JSON")
			return
		}
		if !(handleTrackVolumeInput.Volume >= 0) && !(handleTrackVolumeInput.Volume <= 100) {
			c.JSON(http.StatusBadRequest, "Volume must be between 0 and 100")
			return
		}
		err = client.Volume(ctx, handleTrackVolumeInput.Volume)
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		currentDevice.Volume = handleTrackVolumeInput.Volume
	case "skip":
		track, _ := getNextSong(currentTrackURI)
		playerOpt.URIs = []spotify.URI{track.URI}
		err = client.NextOpt(ctx, &playerOpt)
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, err)
			return
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

	c.JSON(http.StatusAccepted, "Ok: "+action)
}

func pollSpotify() {
	var track Track

	pollingSpotify = true

	fmt.Println("STARTING JUKEBOX WITH DEVICE: " + currentDevice.ID)

	playerState, err := client.PlayerState(context.Background())
	if err != nil {
		fmt.Println("SOMETHING WENT WRONG GETTING PLAYER")
		fmt.Println(err)
	}
	if playerState.Playing {
		fallbackPlaylist.Active = true
		currentTrackURI = playerState.Item.URI
		fmt.Println("CONTINUING SONG: " + playerState.Item.Name + " - " + playerState.Item.Artists[0].Name)
	} else {
		track, _ := getNextSong()
		err := client.PlayOpt(context.Background(), &spotify.PlayOptions{DeviceID: &currentDevice.ID, URIs: []spotify.URI{track.URI}})
		if err != nil {
			fmt.Println("SOMETHING WENT WRONG STARTING PLAYER")
			fmt.Println(err)
		}
		fallbackPlaylist.Active = track.FromFallBackPlaylist
		currentTrackURI = track.URI
		fmt.Println("STARTING SONG: " + track.Name + " - " + track.Artist)
	}

	// Update Current Device
	currentDevice.Active = playerState.Device.Active
	currentDevice.Volume = playerState.Device.Volume

	c := time.Tick(10 * time.Second)

	// Start the main Loop
	for _ = range c {
		var playerState *spotify.PlayerState
		// Check Expiry
		if m, _ := time.ParseDuration("30s"); time.Until(oauthToken.Expiry) < m {
			// Attempt to reAuth
			fmt.Println("OLD TOKEN")
			fmt.Println(oauthToken.AccessToken)
			fmt.Println(oauthToken.RefreshToken)

			client = spotify.New(auth.Client(context.Background(), &oauth2.Token{RefreshToken: oauthToken.RefreshToken}))
			token, err := client.Token()
			if err != nil {
				fmt.Println("SOMETHING WENT WRONG REFRESHING TOKEN")
				fmt.Println(err.Error())
			}

			fmt.Println("NEW TOKEN")
			fmt.Println(token.AccessToken)
			fmt.Println(token.RefreshToken)

			oauthToken.AccessToken = token.AccessToken
			oauthToken.TokenType = token.TokenType
			oauthToken.RefreshToken = token.RefreshToken
			oauthToken.Expiry = token.Expiry
		} else {
			client = spotify.New(auth.Client(context.Background(), &oauth2.Token{AccessToken: oauthToken.AccessToken}))
		}

		// Get Player State
		playerState, err = client.PlayerState(context.Background())
		if err != nil {
			fmt.Println("SOMETHING WENT GETTING PLAYER STATE")
			fmt.Println(err.Error())
		}
		fmt.Println("CURRENT SONG: " + playerState.Item.Name + " - " + playerState.Item.Artists[0].Name)
		fmt.Println("CURRENT PROGRESS: " + strconv.Itoa(playerState.Progress))
		fmt.Println("FALLBACK STATUS: " + strconv.FormatBool(fallbackPlaylist.Active))

		// Update Current Device
		currentDevice.Active = playerState.Device.Active
		currentDevice.Volume = playerState.Device.Volume

		if playerState.Progress == 0 {
			fmt.Println("LOADING NEXT SONG")
			// Remove the track
			if !fallbackPlaylist.Active {
				if fallbackPlaylist.AddToPlaylist {
					// Can't check if song is already in playlist - so just delete it
					_, err := client.RemoveTracksFromPlaylist(context.Background(), fallbackPlaylist.ID, spotify.ID(strings.Replace(string(currentTrackURI), "spotify:track:", "", -1)))
					if err != nil {
						fmt.Println("SOMETHING WENT WRONG REMOVING ITEM FROM PLAYLIST")
						fmt.Println(err)
					}
					fmt.Println("ADDING TRACK TO FALLBACK PLAYLIST: " + currentTrackURI)
					_, err = client.AddTracksToPlaylist(context.Background(), fallbackPlaylist.ID, spotify.ID(strings.Replace(string(currentTrackURI), "spotify:track:", "", -1)))
					if err != nil {
						fmt.Println("SOMETHING WENT WRONG ADDING ITEM TO PLAYLIST")
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
			track, err = getNextSong()
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
				DeviceID: &currentDevice.ID,
				URIs:     []spotify.URI{track.URI},
			}
			err = client.PlayOpt(context.Background(), &playerOpt)

			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func getAllDevices(c *gin.Context) {
	devices, err := client.PlayerDevices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusAccepted, devices)
}

func getCurrentDevice(c *gin.Context) {
	fmt.Println(currentDevice)
	c.JSON(http.StatusAccepted, currentDevice)
}

func setDevice(c *gin.Context) {
	var setDeviceIdInput SetDeviceIdInput
	var device Device

	if err := c.ShouldBindJSON(&setDeviceIdInput); err != nil {
		c.JSON(http.StatusInternalServerError, "Cannot Marshal JSON")
		return
	}
	if setDeviceIdInput.DeviceID == "" {
		c.JSON(http.StatusInternalServerError, "Device ID is required")
		return
	}
	devices, err := client.PlayerDevices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	for _, d := range devices {
		if d.ID == setDeviceIdInput.DeviceID {
			device.Active = d.Active
			device.ID = d.ID.String()
			device.Name = d.Name
			device.Type = d.Type
		}
	}

	if device.ID == "" {
		c.JSON(http.StatusNotFound, "Device Not Found")
		return
	}

	currentDevice.ID = spotify.ID(device.ID)
	currentDevice.Active = false
	currentDevice.Name = device.Name
	currentDevice.Type = device.Type
	c.JSON(http.StatusAccepted, currentDevice)
}
