package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

var currentTrackPlayOptions spotify.PlayOptions

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
		// DEBUG - if no items in queue play fallback

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

		//  DEBUG - Move these into pollSpotify
		setCurrentTrackURI(track.URI)
		setCurrentPlayOptions(spotify.PlayOptions{
			DeviceID: &deviceID,
			URIs:     []spotify.URI{track.URI},
		})
		err = client.PlayOpt(ctx, &currentTrackPlayOptions)

		go pollSpotify(authInput)

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
	fmt.Println("here we go")

	var track Track

	// playerOpt := spotify.PlayOptions{
	// 	DeviceID: &deviceID,
	// }

	// var playerState *spotify.PlayerState

	r := rand.New(rand.NewSource(99))
	c := time.Tick(10 * time.Second)

	for _ = range c {
		client := spotify.New(auth.Client(context.Background(), &authInput))

		//Download the current contents of the URL and do something with it
		fmt.Printf("Grab at %s\n", time.Now())
		playerState, err := client.PlayerState(context.Background())
		fmt.Println(err)
		fmt.Println("CURRENT SONG PROGRESS")
		fmt.Println(playerState.Progress)

		// track, _ := getNextSongByVotes(playerState.Item.URI)
		// Assume if no track - use fallback
		// if err != nil {

		// }
		// if track.URI == "" {
		// 	playerOpt := spotify.PlayOptions{
		// 		DeviceID: &deviceID,
		// 		URIs:     []spotify.URI{track.URI},
		// 	}
		// 	err = client.QueueSongOpt(context.Background(), spotify.ID(strings.Replace(string(track.URI), "spotify:track:", "", -1)), &playerOpt)
		// 	if err != nil {
		// 		fmt.Println(err)
		// 		fmt.Println("QUEUED NEW SONG")
		// 	}
		// }
		// if fallbackPlaylist.Active {
		// 	fmt.Println("fallback active")
		// 	if track.URI == currentTrackPlayOptions.URIs[0] && err != nil {
		// 		fmt.Println(err)
		// 		// DEBUG - hacky - just deleting so wont be able to vote off
		// 		if err := db.Unscoped().Delete(&track).Error; err != nil {
		// 			fmt.Println(err)
		// 		}
		// 		fmt.Println("No Queued Tracks - continuing with fallback playlist")
		// 	} else if track.URI != "" {
		// 		playerOpt := spotify.PlayOptions{
		// 			DeviceID: &deviceID,
		// 			URIs:     []spotify.URI{track.URI},
		// 		}
		// 		err = client.QueueSongOpt(context.Background(), spotify.ID(strings.Replace(string(track.URI), "spotify:track:", "", -1)), &playerOpt)
		// 		if err != nil {
		// 			fmt.Println(err)
		// 			fmt.Println("QUEUED NEW SONG")
		// 		}
		// 	}
		// }

		// If Fallback Playlist is active we need to do a different check
		// This is because there is slim chance the poll interval will match up
		// with the end of the track during the fallback playlist.
		if playerState.Progress == 0 {
			fmt.Println("LOADING NEXT SONG")
			// Remove the track
			// DEBUG - if fallback dont run
			if !fallbackPlaylist.Active {
				fmt.Println("REMOVING TRACK FROM QUEUE: " + currentTrackURI)
				if err := db.First(&track, Track{URI: currentTrackURI}).Error; err != nil {
					fmt.Println(err)
				}
				if err := db.Unscoped().Delete(&track).Error; err != nil {
					fmt.Println(err)
				}
			}
			// Get the next track
			track, err = getNextSongByVotes()
			// DEBUG - assume no more tracks - play backup playlist
			// Send one random track from backup playlist. This is to check for new songs in queue
			if err != nil {
				fallbackPlaylist.Active = true

				// DEBUG - Set Random Offset - currently will only pull first 100 songs. Could set Limit higher?
				// Get Random number for fallback playlist track
				fallBackPlaylist, _ := client.GetPlaylistItems(context.Background(), fallbackPlaylist.ID)

				rand.Seed(time.Now().UnixNano())
				randomPlaylistItem := fallBackPlaylist.Items[(rand.Intn(len(fallBackPlaylist.Items)-1) + 1)]
				setCurrentTrackURI(randomPlaylistItem.Track.Track.URI)

				fmt.Println("No More Tracks - Using fall back playlist")
				fmt.Println("NEXT SONG: " + randomPlaylistItem.Track.Track.Name)

				playerOpt := spotify.PlayOptions{
					DeviceID: &deviceID,
					URIs:     []spotify.URI{randomPlaylistItem.Track.Track.URI},
				}
				err = client.PlayOpt(context.Background(), &playerOpt)
			} else {
				fallbackPlaylist.Active = false
				setCurrentTrackURI(track.URI)

				fmt.Println("NEXT SONG: " + track.Name)

				playerOpt := spotify.PlayOptions{
					DeviceID: &deviceID,
					URIs:     []spotify.URI{track.URI},
				}
				err = client.PlayOpt(context.Background(), &playerOpt)
			}
			if err != nil {
				fmt.Println(err)
			}
		}

		// add a bit of jitter
		jitter := time.Duration(r.Int31n(5000)) * time.Millisecond
		time.Sleep(jitter)

	}
}

// DEBUG - to be implemented
func getCurrentDeviceId(c *gin.Context) {
	c.JSON(http.StatusAccepted, deviceID)
}

func setDeviceId(c *gin.Context) {

}

func setCurrentPlayOptions(playerOpt spotify.PlayOptions) {
	currentTrackPlayOptions = playerOpt
}

func setCurrentTrackURI(uri spotify.URI) {
	currentTrackURI = uri
}
