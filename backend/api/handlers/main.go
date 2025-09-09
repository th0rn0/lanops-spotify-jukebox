package handlers

import (
	"lanops/spotify-jukebox/internal/jukebox"
)

type Client struct {
	jbc *jukebox.Client
}

func New(s *jukebox.Client) Client {
	return Client{
		jbc: s,
	}
}
