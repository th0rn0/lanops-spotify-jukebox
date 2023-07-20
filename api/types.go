package main

import (
	"time"

	"github.com/zmb3/spotify/v2"
	"gorm.io/gorm"
)

type LoginToken struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

type FallbackPlaylist struct {
	URI           spotify.URI
	ID            spotify.ID
	Active        bool
	AddToPlaylist bool
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

type HandleTrackInput struct {
	URI spotify.URI `json:"uri"`
}

type HandleVoteInput struct {
	URI spotify.URI `json:"uri"`
}

type GetSongByUriInput struct {
	URI spotify.URI `json:"uri"`
}

type SetDeviceIdInput struct {
	DeviceId spotify.ID `json:"device_id"`
}

// Models
type Track struct {
	URI                  spotify.URI  `gorm:"primaryKey" json:"uri"`
	Name                 string       `json:"name"`
	Artist               string       `json:"artist"`
	Votes                int64        `json:"votes"`
	FromFallBackPlaylist bool         `gorm:"-" default:"false"`
	Images               []TrackImage `gorm:"foreignKey:TrackURI" json:"images"`
}

// Updating data in same transaction
func (t *Track) BeforeDelete(tx *gorm.DB) (err error) {
	var trackImages []TrackImage
	if err := tx.Where("track_uri = ?", t.URI).Find(&trackImages).Error; err != nil {
		return err
	}
	for _, image := range trackImages {
		tx.Model(&TrackImage{}).Unscoped().Delete(&image)
	}
	return
}

type TrackImage struct {
	ID       uint   `gorm:"primarykey"`
	Height   int    `json:"height"`
	Width    int    `json:"width"`
	URL      string `json:"url"`
	TrackURI spotify.URI
}

// type Device struct {
// 	ID     string `json:"id"`
// 	Name   string `json:"name"`
// 	Type   string `json:"type"`
// 	Active bool   `json:"is_active"`
// }

// type User struct {
// 	gorm.Model
// 	DeviceID string
// 	IP       string
// }
