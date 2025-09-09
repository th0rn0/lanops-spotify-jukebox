package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s Client) AuthCallback(c *gin.Context) {
	token, err := s.jbc.GetAuthTokenFromRequest(c.Request.Context(), c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	if st := c.Request.FormValue("state"); !s.jbc.CheckState(st) {
		c.JSON(http.StatusNotFound, errors.New("State mismatch"))
	}

	if err = s.jbc.Login(token); err != nil {
		c.JSON(http.StatusBadRequest, err)
	}

	c.JSON(http.StatusOK, token)
}

func (s Client) AuthLogin(c *gin.Context) {
	c.JSON(http.StatusOK, s.jbc.GetAuthUrl())
}
