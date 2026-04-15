package jukebox

import (
	"testing"

	"github.com/zmb3/spotify/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	if err := db.AutoMigrate(&BannedTerm{}); err != nil {
		t.Fatalf("migrate BannedTerm: %v", err)
	}
	return db
}

func makeTrack(id, name, album string, artists ...string) *spotify.FullTrack {
	sa := make([]spotify.SimpleArtist, 0, len(artists))
	for _, a := range artists {
		sa = append(sa, spotify.SimpleArtist{Name: a})
	}
	return &spotify.FullTrack{
		SimpleTrack: spotify.SimpleTrack{
			ID:      spotify.ID(id),
			Name:    name,
			Artists: sa,
		},
		Album: spotify.SimpleAlbum{Name: album},
	}
}

func TestCheckFullTrackIsBanned_WordInTrackName(t *testing.T) {
	c := &Client{db: newTestDB(t)}
	if _, err := c.CreateBannedTerm(BannedTerm{Type: "word", Value: "forbidden"}); err != nil {
		t.Fatal(err)
	}
	banned, err := c.CheckFullTrackIsBanned(makeTrack("id1", "This FORBIDDEN Thing", "Album", "Artist"))
	if err != nil {
		t.Fatal(err)
	}
	if !banned {
		t.Fatal("expected banned via case-insensitive track name match")
	}
}

func TestCheckFullTrackIsBanned_WordInAlbumName(t *testing.T) {
	c := &Client{db: newTestDB(t)}
	if _, err := c.CreateBannedTerm(BannedTerm{Type: "word", Value: "bad"}); err != nil {
		t.Fatal(err)
	}
	banned, err := c.CheckFullTrackIsBanned(makeTrack("id1", "Clean", "A Bad Album", "Artist"))
	if err != nil {
		t.Fatal(err)
	}
	if !banned {
		t.Fatal("expected banned via album name match")
	}
}

func TestCheckFullTrackIsBanned_TrackIDExact(t *testing.T) {
	c := &Client{db: newTestDB(t)}
	if _, err := c.CreateBannedTerm(BannedTerm{Type: "track", Value: "abc123"}); err != nil {
		t.Fatal(err)
	}
	banned, err := c.CheckFullTrackIsBanned(makeTrack("abc123", "Clean", "Clean", "Clean"))
	if err != nil {
		t.Fatal(err)
	}
	if !banned {
		t.Fatal("expected banned via exact track ID match")
	}
}

func TestCheckFullTrackIsBanned_TrackIDMismatch(t *testing.T) {
	c := &Client{db: newTestDB(t)}
	if _, err := c.CreateBannedTerm(BannedTerm{Type: "track", Value: "abc123"}); err != nil {
		t.Fatal(err)
	}
	banned, err := c.CheckFullTrackIsBanned(makeTrack("xyz999", "Clean", "Clean", "Clean"))
	if err != nil {
		t.Fatal(err)
	}
	if banned {
		t.Fatal("expected non-matching track ID to pass")
	}
}

func TestCheckFullTrackIsBanned_Clean(t *testing.T) {
	c := &Client{db: newTestDB(t)}
	if _, err := c.CreateBannedTerm(BannedTerm{Type: "word", Value: "forbidden"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.CreateBannedTerm(BannedTerm{Type: "track", Value: "banned-id"}); err != nil {
		t.Fatal(err)
	}
	banned, err := c.CheckFullTrackIsBanned(makeTrack("clean-id", "Clean Song", "Clean Album", "Clean Artist"))
	if err != nil {
		t.Fatal(err)
	}
	if banned {
		t.Fatal("expected clean track to pass")
	}
}

func TestDoesBannedTermExist(t *testing.T) {
	c := &Client{db: newTestDB(t)}
	if _, err := c.CreateBannedTerm(BannedTerm{Type: "word", Value: "term"}); err != nil {
		t.Fatal(err)
	}

	exists, err := c.doesBannedTermExist("term")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("expected stored term to be reported as existing")
	}

	exists, err = c.doesBannedTermExist("nope")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("expected unknown term to be reported as missing")
	}
}
