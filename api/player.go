package main

import (
	"context"
	"fmt"
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

	logger.Info().Msg("Got request for: " + action)

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
	case "stop":
		pollingSpotify = false
		err = client.PauseOpt(ctx, &playerOpt)
		if err != nil {
			logger.Err(err)
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
			logger.Err(err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		currentDevice.Volume = handleTrackVolumeInput.Volume
	case "skip":
		track, _ := getNextSongExcludeURI(currentTrackURI)
		banQuery := c.Query("ban")
		if banQuery == "true" {
			_, err := addBannedTrack(currentTrackURI)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		}
		playerOpt.URIs = []spotify.URI{track.URI}
		err = client.NextOpt(ctx, &playerOpt)
		if err != nil {
			logger.Err(err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		if !fallbackPlaylist.Active {
			var currentTrack Track
			if err := db.First(&currentTrack, Track{URI: currentTrackURI}).Error; err != nil {
				c.JSON(http.StatusNotFound, "Track Not Found")
				return
			}
			if err := db.Unscoped().Delete(&currentTrack).Error; err != nil {
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

	logger.Info().Msg(fmt.Sprintf("STARTING JUKEBOX WITH DEVICE: %s", currentDevice.ID))

	playerState, err := client.PlayerState(context.Background())
	if err != nil {
		logger.Err(err).Msg("SOMETHING WENT WRONG GETTING PLAYER")
	}
	if playerState.Playing {
		fallbackPlaylist.Active = false
		currentTrackURI = playerState.Item.URI
		if err := db.First(&track, Track{URI: currentTrackURI}).Error; err != nil {
			// Assume we cant find the track so must be from fallback playlist
			fallbackPlaylist.Active = true
		}
		logger.Info().Msg("CONTINUING SONG: " + playerState.Item.Name + " - " + playerState.Item.Artists[0].Name)
	} else {
		track, _ := getNextSong()
		err := client.PlayOpt(context.Background(), &spotify.PlayOptions{DeviceID: &currentDevice.ID, URIs: []spotify.URI{track.URI}})
		if err != nil {
			logger.Err(err).Msg("SOMETHING WENT WRONG STARTING PLAYER")
		}
		fallbackPlaylist.Active = track.FromFallBackPlaylist
		currentTrackURI = track.URI
		logger.Info().Msg("STARTING SONG: " + track.Name + " - " + track.Artist)
	}

	// Update Current Device
	currentDevice.Active = playerState.Device.Active
	currentDevice.Volume = playerState.Device.Volume

	c := time.Tick(5 * time.Second)

	// Start the main Loop
	for _ = range c {
		if !pollingSpotify {
			dbDevice := Device{}
			if err := db.First(&dbDevice).Error; err != nil {
				// Assume no Device is Set
				logger.Fatal().Msg("Something went wrong getting device after stopping polling!")
			} else {
				dbDevice.Active = false
				db.Save(&dbDevice)
			}
			break
		}
		var playerState *spotify.PlayerState
		// Check Expiry
		if m, _ := time.ParseDuration("30s"); time.Until(oauthToken.Expiry) < m {
			// Attempt to reAuth
			client = spotify.New(auth.Client(context.Background(), &oauth2.Token{RefreshToken: oauthToken.RefreshToken}))
			token, err := client.Token()
			if err != nil {
				logger.Err(err).Msg("SOMETHING WENT WRONG REFRESHING TOKEN")
			}

			oauthToken.AccessToken = token.AccessToken
			oauthToken.TokenType = token.TokenType
			oauthToken.RefreshToken = token.RefreshToken
			oauthToken.Expiry = token.Expiry

			var dbLoginToken LoginToken

			if err := db.First(&dbLoginToken, LoginToken{}).Error; err == nil {
				if err := db.Unscoped().Delete(&dbLoginToken).Error; err != nil {
					logger.Err(err).Msg("SOMETHING WENT WRONG DELETING OLD TOKEN")
				}
			}

			if err := db.Create(&LoginToken{AccessToken: oauthToken.AccessToken, TokenType: oauthToken.TokenType, RefreshToken: oauthToken.RefreshToken, Expiry: oauthToken.Expiry}).Error; err != nil {
				logger.Err(err).Msg("SOMETHING WENT WRONG SAVING NEW TOKEN")
			}

		} else {
			client = spotify.New(auth.Client(context.Background(), &oauth2.Token{AccessToken: oauthToken.AccessToken}))
		}

		// Get Player State
		playerState, err = client.PlayerState(context.Background())
		if err != nil {
			logger.Err(err).Msg("SOMETHING WENT GETTING PLAYER STATE")
		}
		logger.Info().Msg(fmt.Sprintf("CURRENT SONG: %s - %s", playerState.Item.Name, playerState.Item.Artists[0].Name))
		logger.Info().Msg(fmt.Sprintf("CURRENT PROGRESS: %s / %s", strconv.Itoa(playerState.Progress), strconv.Itoa(playerState.Item.Duration)))
		logger.Info().Msg(fmt.Sprintf("FALLBACK STATUS: %s", strconv.FormatBool(fallbackPlaylist.Active)))

		// Update Current Device
		currentDevice.Active = playerState.Device.Active
		currentDevice.Volume = playerState.Device.Volume

		dbDevice := Device{}
		if err := db.First(&dbDevice).Error; err != nil {
			// Assume no Device is Set
			logger.Fatal().Msg("NO DEVICE SET")
		} else {
			dbDevice.Active = currentDevice.Active
			dbDevice.Volume = currentDevice.Volume
			db.Save(&dbDevice)
		}

		if playerState.Progress == 0 {
			logger.Info().Msg("LOADING NEXT SONG")
			voteToSkipCurrent = voteToSkipDefault
			// Remove the track
			if !fallbackPlaylist.Active {
				if fallbackPlaylist.AddToPlaylist {
					// Can't check if song is already in playlist - so just delete it
					_, err := client.RemoveTracksFromPlaylist(context.Background(), fallbackPlaylist.ID, spotify.ID(strings.Replace(string(currentTrackURI), "spotify:track:", "", -1)))
					if err != nil {
						logger.Err(err).Msg("SOMETHING WENT WRONG REMOVING ITEM FROM PLAYLIST")
					}
					logger.Info().Msg(fmt.Sprintf("ADDING TRACK TO FALLBACK PLAYLIST: %s", currentTrackURI))
					_, err = client.AddTracksToPlaylist(context.Background(), fallbackPlaylist.ID, spotify.ID(strings.Replace(string(currentTrackURI), "spotify:track:", "", -1)))
					if err != nil {
						logger.Err(err).Msg("SOMETHING WENT WRONG ADDING ITEM TO PLAYLIST")
					}

				}
				logger.Info().Msg(fmt.Sprintf("REMOVING TRACK FROM QUEUE: %s", currentTrackURI))
				if err := db.First(&track, Track{URI: currentTrackURI}).Error; err != nil {
					logger.Err(err)
				}
				if err := db.Unscoped().Delete(&track).Error; err != nil {
					logger.Err(err)
				}
			}
			// Get the next track
			track, err = getNextSong()
			if err != nil {
				logger.Err(err).Msg("SOMETHING WENT WRONG GETTING NEXT SONG")
			}
			currentTrackURI = track.URI
			fallbackPlaylist.Active = track.FromFallBackPlaylist

			if fallbackPlaylist.Active {
				logger.Info().Msg("No More Tracks - Using fall back playlist")
			}

			logger.Info().Msg("NEXT SONG: " + track.Name + " - " + track.Artist)

			playerOpt := spotify.PlayOptions{
				DeviceID: &currentDevice.ID,
				URIs:     []spotify.URI{track.URI},
			}
			err = client.PlayOpt(context.Background(), &playerOpt)

			if err != nil {
				logger.Err(err)
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
	logger.Info().Msg(fmt.Sprintf("%s", currentDevice))
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
			device.ID = d.ID
			device.Name = d.Name
			device.Type = d.Type
		}
	}

	if device.ID == "" {
		c.JSON(http.StatusNotFound, "Device Not Found")
		return
	}

	currentDevice.ID = device.ID
	currentDevice.Name = device.Name
	currentDevice.Active = device.Active
	currentDevice.Type = device.Type

	var dbDevice Device

	if err := db.First(&dbDevice, Device{}).Error; err == nil {
		if err := db.Unscoped().Delete(&dbDevice).Error; err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
	}

	if err := db.Create(&Device{ID: currentDevice.ID, Name: currentDevice.Name, Active: currentDevice.Active, Type: currentDevice.Type}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusAccepted, currentDevice)
}
