package handlers

import (
	"errors"
	"lanops/spotify-jukebox/internal/jukebox"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify/v2"
	"gorm.io/gorm"
)

func (s Client) GetTracks(c *gin.Context) {
	var tracks []jukebox.Track
	tracks, err := s.jbc.GetTracks()
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	}
	c.JSON(http.StatusOK, tracks)
}

type AddTrackInput struct {
	Id spotify.ID `json:"id"`
}

func (s Client) AddTrack(c *gin.Context) {
	var addTrackInput AddTrackInput
	if err := c.ShouldBindJSON(&addTrackInput); err != nil {
		c.JSON(http.StatusBadRequest, "Cannot Marshal JSON")
		return
	}
	track, err := s.jbc.GetFullTrackFromSpotify(spotify.ID(addTrackInput.Id))
	if err != nil {
		c.JSON(http.StatusBadRequest, "Cannot find song")
		return
	}

	if _, err := s.jbc.AddTrackToQueue(track); err != nil {
		if !errors.Is(err, jukebox.ErrTrackBanned) && !errors.Is(err, gorm.ErrCheckConstraintViolated) {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		if errors.Is(err, jukebox.ErrTrackBanned) {
			c.JSON(http.StatusBadRequest, "Song is banned or contains banned words")
			return
		}
		c.JSON(http.StatusBadRequest, "Song already in queue")
		return
	}

	c.JSON(http.StatusAccepted, track)
}

func (s Client) GetCurrentTrack(c *gin.Context) {
	c.JSON(http.StatusOK, s.jbc.GetCurrentTrack())
}

type GetTrackByIdInput struct {
	Id spotify.ID `json:"id"`
}

func (s Client) GetTrackById(c *gin.Context) {
	var getTrackByIdInput GetTrackByIdInput
	if err := c.ShouldBindJSON(&getTrackByIdInput); err != nil {
		c.JSON(http.StatusBadRequest, "Cannot Marshal JSON")
		return
	}
	track, err := s.jbc.GetTrackFromQueueById(spotify.ID(getTrackByIdInput.Id))
	if err != nil {
		c.JSON(http.StatusBadRequest, "Cannot find song")
		return
	}
	c.JSON(http.StatusOK, track)
}

type DeleteTrackByIdInput struct {
	Id spotify.ID `json:"id"`
}

func (s Client) DeleteTrackById(c *gin.Context) {
	var deleteTrackByIdInput DeleteTrackByIdInput
	if err := c.ShouldBindJSON(&deleteTrackByIdInput); err != nil {
		c.JSON(http.StatusBadRequest, "Cannot Marshal JSON")
		return
	}
	err := s.jbc.DeleteTrackFromQueueById(spotify.ID(deleteTrackByIdInput.Id))
	if err != nil {
		// TODO proper response on no record
		c.JSON(http.StatusBadRequest, "Cannot find song")
		return
	}
	c.JSON(http.StatusOK, true)
}
