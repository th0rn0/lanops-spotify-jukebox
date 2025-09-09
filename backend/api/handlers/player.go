package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s Client) PlayerStart(c *gin.Context) {
	s.jbc.SetActive(true)
	s.jbc.SetPaused(false)
	c.JSON(http.StatusOK, true)
}

func (s Client) PlayerStop(c *gin.Context) {
	s.jbc.SetActive(false)
	c.JSON(http.StatusOK, false)
}

type PlayerVolumeInput struct {
	Volume int `json:"volume"`
}

func (s Client) PlayerVolume(c *gin.Context) {
	var playerVolumeInput PlayerVolumeInput
	if err := c.ShouldBindJSON(&playerVolumeInput); err != nil {
		c.JSON(http.StatusInternalServerError, "Cannot Marshal JSON")
		return
	}
	if !(playerVolumeInput.Volume >= 0) && !(playerVolumeInput.Volume <= 100) {
		c.JSON(http.StatusBadRequest, "Volume must be between 0 and 100")
		return
	}
	if err := s.jbc.SetVolume(playerVolumeInput.Volume); err != nil {
		c.JSON(http.StatusInternalServerError, "Something went wrong")
	}
	c.JSON(http.StatusOK, true)
}

func (s Client) PlayerSkip(c *gin.Context) {
	s.jbc.SetSkip(true)
	c.JSON(http.StatusOK, true)
}

func (s Client) PlayerPause(c *gin.Context) {
	s.jbc.SetPaused(true)
	c.JSON(http.StatusOK, true)
}
