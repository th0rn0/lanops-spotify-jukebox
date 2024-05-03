package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/zmb3/spotify/v2"
)

func getNextSong() (Track, error) {
	var nextTrack Track
	var err error
	if voteToSkipEnabled {
		nextTrack, err = getNextSongRandom()
	} else {
	nextTrack, err = getNextSongByVotes()
	}
	if err != nil {
		// Assume no record - get from fallback playlist
		nextTrack = assignFallback(nextTrack)
	} else {
		fallbackPlaylist.Active = false
	}
	return nextTrack, nil
}

func getNextSongExcludeURI(excludeUri spotify.URI) (Track, error) {
	var nextTrack Track
	var err error
	nextTrack, err = getNextSongByVotesExcludeURI(excludeUri)
	if err != nil {
		// Assume no record - get from fallback playlist
		nextTrack = assignFallback(nextTrack)
	} else {
		fallbackPlaylist.Active = false
	}
	return nextTrack, nil
}

func getNextSongRandom() (Track, error) {
	var track Track
	if err := db.Raw("SELECT * FROM tracks ORDER BY random()").First(&track).Error; err != nil {
		return track, err
	}
	return track, nil
}

func getNextSongByVotes() (Track, error) {
	var track Track
	if err := db.Raw("SELECT * FROM tracks WHERE votes = ( SELECT MAX(votes) FROM tracks )").First(&track).Error; err != nil {
		return track, err
	}
	return track, nil
}

func getNextSongByVotesExcludeURI(excludeUri spotify.URI) (Track, error) {
	var track Track
	if err := db.Raw("SELECT * FROM tracks WHERE votes = ( SELECT MAX(votes) FROM tracks ) AND uri != ?", excludeUri).First(&track).Error; err != nil {
		return track, err
	}
	return track, nil
}

func getRandomFallbackPlaylistItem() spotify.PlaylistItem {
	fallBackPlaylist, _ := client.GetPlaylistItems(context.Background(), fallbackPlaylist.ID)

	// Get playlist again with a limit of 1 and random offset between 1 and the total of tracks in the playlist
	rand.Seed(time.Now().UnixNano())
	randomOffset := rand.Intn(fallBackPlaylist.Total-1) + 1
	fallBackPlaylist, _ = client.GetPlaylistItems(context.Background(), fallbackPlaylist.ID, spotify.Limit(1), spotify.Offset(randomOffset))

	return fallBackPlaylist.Items[0]
}

func assignFallback(nextTrack Track) Track {
	nextSongFromPlayList := getRandomFallbackPlaylistItem()
	nextTrack.Artist = nextSongFromPlayList.Track.Track.Artists[0].Name
	nextTrack.Name = nextSongFromPlayList.Track.Track.Name
	nextTrack.URI = nextSongFromPlayList.Track.Track.URI
	nextTrack.FromFallBackPlaylist = true

	fallbackPlaylist.Active = true

	for _, trackImage := range nextSongFromPlayList.Track.Track.Album.Images {

		nextTrack.Images = append(nextTrack.Images, TrackImage{
			Height: trackImage.Height,
			Width:  trackImage.Width,
			URL:    trackImage.URL,
		})
	}

	return nextTrack
}
