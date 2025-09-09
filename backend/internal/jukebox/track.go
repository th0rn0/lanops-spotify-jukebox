package jukebox

import (
	"context"
	"errors"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/exp/rand"
	"gorm.io/gorm"
)

type Track struct {
	Id               spotify.ID   `gorm:"primaryKey" json:"id"`
	Name             string       `json:"name"`
	Artist           string       `json:"artist"`
	FallbackPlaylist bool         `gorm:"-" default:"false"`
	Images           []TrackImage `gorm:"foreignKey:TrackId" json:"images"`
}

type TrackImage struct {
	ID      uint            `gorm:"primarykey"`
	Height  spotify.Numeric `json:"height"`
	Width   spotify.Numeric `json:"width"`
	URL     string          `json:"url"`
	TrackId spotify.ID
}

func (t *Track) BeforeDelete(tx *gorm.DB) (err error) {
	var trackImages []TrackImage
	if err := tx.Where("track_id = ?", t.Id).Find(&trackImages).Error; err != nil {
		return err
	}
	for _, image := range trackImages {
		tx.Model(&TrackImage{}).Unscoped().Delete(&image)
	}
	return
}

func (c *Client) getNext() (track Track, err error) {
	track, err = c.getNextRandomFromQueue()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		track, err = c.getNextFromFallbackPlaylist()
	}
	if err != nil {
		return track, err
	}
	return track, nil
}

func (c *Client) getNextFromFallbackPlaylist() (track Track, err error) {

	fallBackPlaylist, err := c.spotify.client.GetPlaylistItems(context.Background(), c.fallbackPlaylistId)
	if err != nil {
		return track, err
	}
	randomOffset := rand.Intn(len(fallBackPlaylist.Items))

	track.Artist = fallBackPlaylist.Items[randomOffset].Track.Track.Artists[0].Name
	track.Name = fallBackPlaylist.Items[randomOffset].Track.Track.Name
	track.Id = fallBackPlaylist.Items[randomOffset].Track.Track.ID
	track.FallbackPlaylist = true

	for _, trackImage := range fallBackPlaylist.Items[randomOffset].Track.Track.Album.Images {
		track.Images = append(track.Images, TrackImage{
			Height: trackImage.Height,
			Width:  trackImage.Width,
			URL:    trackImage.URL,
		})
	}

	return track, nil
}

func (c *Client) getNextRandomFromQueue() (track Track, err error) {
	if err = c.db.Raw("SELECT * FROM tracks ORDER BY random()").First(&track).Error; err != nil {
		return track, err
	}
	return track, nil
}

func (c *Client) addCurrentTrackToFallbackPlaylist() error {
	// Can't check if song is already in playlist - so just delete it
	_, err := c.spotify.client.RemoveTracksFromPlaylist(context.Background(), c.fallbackPlaylistId, spotify.ID(c.current.Id))
	if err != nil {
		return err
	}
	_, err = c.spotify.client.AddTracksToPlaylist(context.Background(), c.fallbackPlaylistId, c.current.Id)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) deleteCurrentTrackFromQueue() error {
	if err := c.db.First(&Track{}, c.current.Id).Error; err != nil {
		return err
	}
	if err := c.db.Unscoped().Delete(&c.current).Error; err != nil {
		return err
	}
	c.current = Track{}
	return nil
}

func (c *Client) GetTracks() (tracks []Track, err error) {
	if err = c.db.Preload("Images").Find(&tracks).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return tracks, err
	}
	return tracks, nil
}

func (c *Client) GetCurrentTrack() (track Track) {
	return c.current
}

func (c *Client) GetTrackFromQueueById(id spotify.ID) (track Track, err error) {
	if err := c.db.Preload("Images").First(&track, Track{Id: id}).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return track, err
	}
	return track, nil
}

func (c *Client) GetFullTrackFromSpotify(id spotify.ID) (track *spotify.FullTrack, err error) {
	track, err = c.spotify.client.GetTrack(context.Background(), id)
	return track, err
}

func (c *Client) AddTrackToQueue(fullTrack *spotify.FullTrack) (track Track, err error) {
	// Check track is not banned or contains banned words
	bannedFlag, err := c.CheckFullTrackIsBanned(fullTrack)
	if err != nil {
		return track, err
	}
	if bannedFlag {
		return track, ErrTrackBanned
	}
	// Get Track Images
	trackImages := []TrackImage{}
	for _, image := range fullTrack.Album.Images {
		thisImage := TrackImage{
			URL:     image.URL,
			Height:  image.Height,
			Width:   image.Width,
			TrackId: fullTrack.ID,
		}
		trackImages = append(trackImages, thisImage)
	}
	track = Track{
		Id:     fullTrack.ID,
		Name:   fullTrack.Name,
		Artist: fullTrack.Artists[0].Name,
		Images: trackImages,
	}
	if err := c.db.Create(track).Error; err != nil {
		return track, err
	}
	return track, nil
}

func (c *Client) DeleteTrackFromQueueById(id spotify.ID) (err error) {
	if c.current.Id == id {
		c.SetSkip(true)
	}
	if err := c.db.Unscoped().Delete(&id).Error; err != nil {
		return err
	}
	return nil
}
