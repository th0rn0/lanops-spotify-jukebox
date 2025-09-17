package jukebox

import (
	"context"

	"github.com/zmb3/spotify/v2"
)

func (c *Client) SetFallbackPlaylist(id string) {
	c.fallbackPlaylistId = spotify.ID(id)
}

func (c *Client) SetActive(state bool) {
	c.active = state
}

func (c *Client) SetVolume(volume int) (err error) {
	if err = c.spotify.client.Volume(context.Background(), volume); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetVolume() (volume int, err error) {
	playerState, err := c.spotify.client.PlayerState(context.Background())
	if err != nil {
		return volume, err
	}
	return int(playerState.Device.Volume), nil
}

func (c *Client) SetSkip(state bool) {
	c.skip.active = state
}

func (c *Client) SetPaused(state bool) {
	c.paused = state
}
