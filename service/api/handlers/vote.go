package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s Client) VoteToSkip(c *gin.Context) {
	s.jbc.VoteToSkip()
	c.JSON(http.StatusAccepted, true)
}
