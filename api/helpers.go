package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/zmb3/spotify/v2"
)

func getNextSong(excludeUri ...spotify.URI) (Track, error) {
	var nextTrack Track
	var err error

	if len(excludeUri) > 0 {
		nextTrack, err = getNextSongByVotes(excludeUri[0])
	} else {
		nextTrack, err = getNextSongByVotes()
	}
	// DEBUG - assume no record - get from fallback playlist
	// DEBUG - fix this
	if err != nil {
		nextSongFromPlayList := getRandomFallbackPlaylistItem()
		// if err != nil {
		// 	return nextTrack, err
		// }
		nextTrack.Artist = nextSongFromPlayList.Track.Track.Artists[0].Name
		nextTrack.Name = nextSongFromPlayList.Track.Track.Name
		nextTrack.URI = nextSongFromPlayList.Track.Track.URI
		nextTrack.FromFallBackPlaylist = true

		for _, trackImage := range nextSongFromPlayList.Track.Track.Album.Images {

			nextTrack.Images = append(nextTrack.Images, TrackImage{
				Height: trackImage.Height,
				Width:  trackImage.Width,
				URL:    trackImage.URL,
			})
		}
	}
	return nextTrack, nil
}

func getNextSongByVotes(excludeUri ...spotify.URI) (Track, error) {
	var track Track

	if len(excludeUri) > 0 {
		if err := db.Raw("SELECT * FROM tracks WHERE votes = ( SELECT MAX(votes) FROM tracks ) AND uri != ?", excludeUri[0]).First(&track).Error; err != nil {
			return track, err
		}
	} else {
		if err := db.Raw("SELECT * FROM tracks WHERE votes = ( SELECT MAX(votes) FROM tracks )").First(&track).Error; err != nil {
			return track, err
		}
	}
	return track, nil
}

func getRandomFallbackPlaylistItem() spotify.PlaylistItem {
	// DEBUG - Set Random Offset - currently will only pull first 100 songs. Could set Limit higher?
	// Get Random number for fallback playlist track
	// We add a single track so that we can still check playerState.Progress == 0
	fallBackPlaylist, _ := client.GetPlaylistItems(context.Background(), fallbackPlaylist.ID)

	rand.Seed(time.Now().UnixNano())
	return fallBackPlaylist.Items[(rand.Intn(len(fallBackPlaylist.Items)-1) + 1)]
}
