package jukebox

import (
	"context"
	"fmt"
	"time"

	"github.com/zmb3/spotify/v2"
)

func (c *Client) Run() error {
	// Check we have a login before continuing
	c.log.Info().Msg("Checking for Valid Spotify token...")
	for {
		time.Sleep(5 * time.Second)

		if c.hasAuthToken() {
			c.log.Info().Msg("Spotify Token found - Attempting Login")
			if err := c.Login(c.spotify.token); err != nil {
				c.log.Err(err).Msg("Error attempting login in to Client with saved token. Please login again")
			}
		}

		if c.spotify.token != nil && c.spotify.client != nil {

			token, err := c.spotify.client.Token()

			if err == nil && token != nil {
				c.log.Info().Msg("Valid Spotify token found! Starting Spotify polling!")
				break
			}
		}
	}

	// Start Main Loop
	for {
		// Wait Time
		time.Sleep(5 * time.Second)

		// Reauth Token
		if m, _ := time.ParseDuration("30s"); time.Until(c.spotify.token.Expiry) < m {
			if err := c.Login(c.spotify.token); err != nil {
				c.log.Err(err).Msg("Couldn't refresh Token")
			}
		}

		if c.active {
			if c.paused {
				c.spotify.client.Pause(context.Background())
				for {
					if !c.paused {
						break
					}
				}
				c.spotify.client.Play(context.Background())
			}
			// Reset the PlayerState
			playerState, err := c.spotify.client.PlayerState(context.Background())
			if err != nil {
				c.log.Err(err).Msg("SOMETHING WENT WRONG GETTING PLAYER STATE")
			}
			if playerState.Device.Active {
				// Sync PlayerState Item with current
				if playerState.Playing {
					c.current.Artist = playerState.Item.Artists[0].Name
					c.current.Name = playerState.Item.Name
					c.current.Id = playerState.Item.ID
					c.current.Images = []TrackImage{}
					for _, trackImage := range playerState.Item.Album.Images {
						c.current.Images = append(c.current.Images, TrackImage{
							Height: trackImage.Height,
							Width:  trackImage.Width,
							URL:    trackImage.URL,
						})
					}
					c.log.Info().Msg(fmt.Sprintf("CURRENT SONG: %s - %s", playerState.Item.Name, playerState.Item.Artists[0].Name))
					c.log.Info().Msg(fmt.Sprintf("CURRENT PROGRESS: %d / %d", playerState.Progress, playerState.Item.Duration))
					c.log.Info().Msg(fmt.Sprintf("FALLBACK STATUS: %t", c.current.FallbackPlaylist))
				}
				// RESET THE PLAYER - Maybe move this outside the loop. Realistically the spotify account will ONLY be used for the jukebox.
				// However I've tried to make sure the jukebox can reliably start, stop and be left running all event
				// Cant seem to reliably reset the queue - Because of how the polling works, If there are items we will assume that the loop has just started.
				// Therefore we will stop the player to allow a fresh playOpt event to fire.
				// This stops spotify taking over from the jukebox
				queue, _ := c.spotify.client.GetQueue(context.Background())
				// For what ever reason, the items in the queue for ONE SONG is 10
				if len(queue.Items) > 10 {
					c.spotify.client.Pause(context.Background())
				}
				c.spotify.client.Repeat(context.Background(), "off")
				c.spotify.client.Shuffle(context.Background(), false)

				// Process the Next Track
				if (!playerState.Playing || playerState.Progress == 0) || c.shouldSkip() {
					if c.shouldSkip() {
						c.log.Info().Msg("Skip triggered")
					}
					c.log.Info().Msg("Getting Next Track")
					if !c.current.FallbackPlaylist {
						if !c.shouldSkip() {
							if err := c.addCurrentTrackToFallbackPlaylist(); err != nil {
								c.log.Err(err).Msg("Something went wrong adding a track to the fallback playlist")
							}
						}
						if err := c.deleteCurrentTrackFromQueue(); err != nil {
							c.log.Err(err).Msg("Something went wrong deleting track From the queue")
						}
					}
					// Get New Track
					c.current, err = c.getNext()
					c.log.Info().Msg(fmt.Sprintf("FALLBACK STATUS: %t", c.current.FallbackPlaylist))

					if err != nil {
						return err
					}
					playerOpt := spotify.PlayOptions{
						URIs: []spotify.URI{spotify.URI(fmt.Sprintf("spotify:track:%s", c.current.Id))},
					}

					err = c.spotify.client.PlayOpt(context.Background(), &playerOpt)
					if err != nil {
						c.log.Err(err).Msg("Something went wrong trying to play next song")
					}
					// Reset Skip
					c.resetSkip()
				}
			} else {
				c.log.Info().Msg("No Active device found. Please set a device.")
			}
		} else {
			// Stop Spotify playback if it isn't already stopped
			c.current = Track{}
			c.current.FallbackPlaylist = true
			playerState, err := c.spotify.client.PlayerState(context.Background())
			if err != nil {
				c.log.Err(err).Msg("SOMETHING WENT WRONG GETTING PLAYER STATE")
			}
			if playerState.Playing {
				c.spotify.client.Pause(context.Background())
			}
		}
	}
}
