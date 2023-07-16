package main

import (
	"github.com/zmb3/spotify/v2"
)

type LoginToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Expiry       string `json:"expiry"`
}

type FallbackPlaylist struct {
	URI    spotify.URI
	ID     spotify.ID
	Active bool
}

// Outputs
type SearchOutput struct {
	Playlists []PlaylistSearchOutput `json:"playlist"`
	Artists   []ArtistSearchOutput   `json:"artist"`
	Tracks    []TrackSearchOutput    `json:"track"`
}

type PlaylistSearchOutput struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type ArtistSearchOutput struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type TrackSearchOutput struct {
	Name   string          `json:"name"`
	Artist string          `json:"artist"`
	ID     string          `json:"id"`
	Images []spotify.Image `json:"images"`
}

type GetDeviceIdsOutput struct {
	PlayerDevices []spotify.PlayerDevice `json:"devices"`
}

type GetTrackOutput struct {
	URI    spotify.URI     `json:"uri"`
	Name   string          `json:"name"`
	Artist string          `json:"artist"`
	Votes  int64           `json:"votes"`
	Images []spotify.Image `json:"images"`
}

// Inputs
type HandlePlayerInput struct {
	URI spotify.URI `json:"uri"`
}

type HandleSongInput struct {
	URI spotify.URI `json:"uri"`
}

type HandleVoteInput struct {
	URI spotify.URI `json:"uri"`
}

type GetSongByUriInput struct {
	URI spotify.URI `json:"uri"`
}

// Models
type Track struct {
	URI    spotify.URI  `gorm:"primaryKey" json:"uri"`
	Name   string       `json:"name"`
	Artist string       `json:"artist"`
	Votes  int64        `json:"votes"`
	Images []TrackImage `gorm:"foreignKey:URI;references:URI" json:"images"`
}

// DEBUG - remove images on cascade
type TrackImage struct {
	ID     uint   `gorm:"primarykey"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
	URL    string `json:"url"`
	URI    spotify.URI
}

type Device struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Active bool   `json:"is_active"`
}
