package jukebox

import (
	"context"

	"github.com/zmb3/spotify/v2"
)

func (c *Client) Search(ctx context.Context, searchTerm string) (results *spotify.SearchResult, err error) {
	results, err = c.spotify.client.Search(ctx, searchTerm, spotify.SearchTypeTrack)
	if err != nil {
		return results, err
	}
	return results, nil
}
