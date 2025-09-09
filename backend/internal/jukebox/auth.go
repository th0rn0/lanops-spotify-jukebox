package jukebox

import (
	"context"
	"errors"
	"net/http"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

func (c *Client) GetAuthUrl() string {
	return c.spotify.auth.AuthURL(state)
}

func (c *Client) GetAuthTokenFromRequest(ctx context.Context, req *http.Request) (token *oauth2.Token, err error) {
	return c.spotify.auth.Token(ctx, state, req)
}

func (c *Client) CheckState(st string) bool {
	return state == st
}

func (c *Client) getAuthToken() (*oauth2.Token, error) {
	if c.spotify.token != nil {
		return c.spotify.token, nil
	}
	token, err := c.getAuthTokenFromDb()
	if err != nil {
		return c.spotify.token, err
	}
	return token, err
}

func (c *Client) setAuthToken(token *oauth2.Token, saveToDb bool) error {
	c.spotify.token = token
	if saveToDb {
		if err := c.saveAuthTokenToDb(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) hasAuthToken() bool {
	if c.spotify.token == nil {
		return false
	}
	return true
}

func (c *Client) Login(token *oauth2.Token) error {
	c.spotify.client = spotify.New(c.spotify.auth.Client(context.TODO(), token))
	if _, err := c.spotify.client.CurrentUser(context.TODO()); err != nil {
		return err
	}
	token, err := c.spotify.client.Token()
	if err != nil {
		return err
	}
	c.setAuthToken(token, true)
	return nil
}

func (c *Client) saveAuthTokenToDb() error {
	var oldToken *oauth2.Token
	if err := c.db.First(&oldToken).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if oldToken.AccessToken != "" {
		if err := c.db.Where("access_token = ?", oldToken.AccessToken).Delete(&oldToken).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	if err := c.db.Create(&c.spotify.token).Error; err != nil {
		return err
	}
	return nil
}

func (c *Client) getAuthTokenFromDb() (*oauth2.Token, error) {
	var token *oauth2.Token
	if err := c.db.First(&token).Error; err != nil {
		return token, err
	}
	return token, nil
}
