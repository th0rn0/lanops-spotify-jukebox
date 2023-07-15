package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func serveLoginLink(c *gin.Context) {
	url := auth.AuthURL(state)
	c.JSON(http.StatusOK, url)
}

func handleAuth(c *gin.Context) {
	tok, err := auth.Token(c.Request.Context(), state, c.Request)
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, err)
	}
	if st := c.Request.FormValue("state"); st != state {
		log.Fatalf("State mismatch: %s != %s\n", st, state)
		c.JSON(http.StatusNotFound, err)
	}
	// Return Auth to client
	c.JSON(http.StatusOK, LoginToken{
		AccessToken:  tok.AccessToken,
		TokenType:    tok.TokenType,
		RefreshToken: tok.RefreshToken,
		Expiry:       tok.Expiry.String(),
	})
}
