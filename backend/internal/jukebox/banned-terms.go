package jukebox

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/zmb3/spotify/v2"
)

var ErrTrackBanned = errors.New("Track is banned")

type BannedTerm struct {
	Type  string
	Value string
}

func (c *Client) LoadBannedTermsFromFile() error {
	bannedTracksFile, err := os.Open(c.cfg.BannedTerms.TracksFileLocation)
	if err != nil {
		return err
	}
	defer bannedTracksFile.Close()

	scannerBannedTracks := bufio.NewScanner(bannedTracksFile)
	for scannerBannedTracks.Scan() {
		isBanned, err := c.doesBannedTermExist(scannerBannedTracks.Text())
		if err != nil {
			return err
		}
		if !isBanned {
			_, err := c.CreateBannedTerm(BannedTerm{Type: "track", Value: scannerBannedTracks.Text()})
			if err != nil {
				return err
			}
		}
	}

	if err := scannerBannedTracks.Err(); err != nil {
		return err
	}

	bannedWordsFile, err := os.Open(c.cfg.BannedTerms.WordsFileLocation)
	if err != nil {
		return err
	}
	defer bannedWordsFile.Close()

	scannerBannedWords := bufio.NewScanner(bannedWordsFile)
	for scannerBannedWords.Scan() {
		isBanned, err := c.doesBannedTermExist(scannerBannedWords.Text())
		if err != nil {
			return err
		}
		if !isBanned {
			_, err := c.CreateBannedTerm(BannedTerm{Type: "word", Value: strings.ToLower(scannerBannedWords.Text())})
			if err != nil {
				return err
			}
		}
	}

	if err := scannerBannedWords.Err(); err != nil {
		return err
	}
	return nil
}

func (c *Client) CreateBannedTerm(bannedTerm BannedTerm) (BannedTerm, error) {
	result := c.db.Create(&bannedTerm)
	if result.Error != nil {
		return bannedTerm, result.Error
	}
	return bannedTerm, nil
}

func (c *Client) CheckFullTrackIsBanned(fullTrack *spotify.FullTrack) (bool, error) {
	var bannedTerms []BannedTerm
	if err := c.db.Find(&bannedTerms).Error; err != nil {
		return true, err
	}
	for _, bannedTerm := range bannedTerms {
		if bannedTerm.Type == "word" {
			fmt.Println("we are checking")
			if strings.Contains(strings.ToLower(fullTrack.Name), bannedTerm.Value) || strings.Contains(strings.ToLower(fullTrack.Album.Name), bannedTerm.Value) {
				return true, nil
			}
			for _, artist := range fullTrack.Artists {
				if strings.Contains(artist.Name, bannedTerm.Value) {
					return true, nil
				}
			}
		}
		if bannedTerm.Type == "track" && bannedTerm.Value == fullTrack.ID.String() {
			return true, nil
		}
	}
	return false, nil
}

func (c *Client) doesBannedTermExist(searchTerm string) (bool, error) {
	var count int64
	if err := c.db.Model(BannedTerm{}).Where("value = ?", strings.ToLower(searchTerm)).Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}
