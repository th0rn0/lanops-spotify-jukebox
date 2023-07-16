package main

import "github.com/zmb3/spotify/v2"

func getNextSongByVotes(excludeUri ...spotify.URI) (Track, error) {
	var track Track

	// if err := db.Raw("SELECT MAX(votes) FROM tracks").First(&track).Error; err != nil {
	if excludeUri == nil {
		if err := db.Raw("SELECT * FROM tracks WHERE votes = ( SELECT MAX(votes) FROM tracks )").First(&track).Error; err != nil {
			return track, err
		}
	} else {
		if err := db.Raw("SELECT * FROM tracks WHERE votes = ( SELECT MAX(votes) FROM tracks ) AND uri != ?", excludeUri).First(&track).Error; err != nil {
			return track, err
		}
	}
	return track, nil
}
