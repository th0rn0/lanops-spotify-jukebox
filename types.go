package main

import "github.com/zmb3/spotify/v2"

type LoginToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Expiry       string `json:"expiry"`
}

// Outputs
type SearchResult struct {
	PlaylistResult PlaylistResult `json:"playlist"`
	ArtistResult   ArtistResult   `json:"artist"`
	TrackResult    TrackResult    `json:"track"`
}

type PlaylistResult struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type ArtistResult struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type TrackResult struct {
	Name   string `json:"name"`
	Artist string `json:"artist"`
	ID     string `json:"id"`
}

// Inputs
type PlayerSongInput struct {
	URI spotify.URI `json:"uri"`
}
