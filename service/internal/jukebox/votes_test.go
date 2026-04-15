package jukebox

import (
	"testing"

	"lanops/spotify-jukebox/internal/config"
)

func TestShouldSkip_Initial(t *testing.T) {
	c := &Client{cfg: config.Config{VoteCountToSkip: 3}}
	if c.shouldSkip() {
		t.Fatal("fresh client should not be marked for skip")
	}
}

func TestShouldSkip_BelowThreshold(t *testing.T) {
	c := &Client{cfg: config.Config{VoteCountToSkip: 3}}
	c.VoteToSkip()
	c.VoteToSkip()
	if c.shouldSkip() {
		t.Fatal("2 votes below threshold 3 must not trigger skip")
	}
}

func TestShouldSkip_AtThreshold(t *testing.T) {
	c := &Client{cfg: config.Config{VoteCountToSkip: 3}}
	c.VoteToSkip()
	c.VoteToSkip()
	c.VoteToSkip()
	if !c.shouldSkip() {
		t.Fatal("3 votes at threshold 3 must trigger skip")
	}
}

func TestShouldSkip_ActiveFlagShortCircuits(t *testing.T) {
	c := &Client{cfg: config.Config{VoteCountToSkip: 999}}
	c.SetSkip(true)
	if !c.shouldSkip() {
		t.Fatal("SetSkip(true) must trigger skip regardless of vote count")
	}
}

func TestResetSkip_ClearsBoth(t *testing.T) {
	c := &Client{cfg: config.Config{VoteCountToSkip: 2}}
	c.SetSkip(true)
	c.VoteToSkip()
	c.VoteToSkip()
	c.resetSkip()
	if c.shouldSkip() {
		t.Fatal("resetSkip must clear both the active flag and the vote count")
	}
}
