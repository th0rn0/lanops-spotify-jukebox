package main

import "github.com/zmb3/spotify/v2"

func containsBannedWord(word string) bool {
	var bannedWords []BannedWord

	result := db.Where("word LIKE ?", "%"+word+"%").Find(&bannedWords)

	if result.Error != nil {
		return false
	}
	if len(bannedWords) == 0 {
		return false
	}
	return true
}

func addBannedWord(word string) (BannedWord, error) {
	bannedWord := BannedWord{Word: word}
	result := db.Create(&bannedWord)
	if result.Error != nil {
		return bannedWord, result.Error
	}
	return bannedWord, nil
}

func isBannedTrack(trackUri spotify.URI) bool {
	var bannedTracks []BannedTrack

	result := db.Where("track_uri LIKE ?", "%"+trackUri+"%").Find(&bannedTracks)

	if result.Error != nil {
		return false
	}
	if len(bannedTracks) == 0 {
		return false
	}
	return true
}

func addBannedTrack(trackUri spotify.URI) (BannedTrack, error) {
	bannedTrack := BannedTrack{TrackURI: trackUri}
	result := db.Create(&bannedTrack)
	if result.Error != nil {
		return bannedTrack, result.Error
	}
	return bannedTrack, nil
}
