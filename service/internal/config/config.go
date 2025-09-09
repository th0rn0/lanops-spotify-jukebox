package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func Load() Config {
	godotenv.Load()

	// DB
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		log.Fatal("❌ DB_PATH not set in environment")
	}

	// BANNED TERMS
	bannedTermsTracksFileLocation := os.Getenv("BANNED_TERMS_TRACKS_FILE_LOCATION")
	if bannedTermsTracksFileLocation == "" {
		log.Fatal("❌ BANNED_TERMS_TRACKS_FILE_LOCATION not set in environment")
	}
	bannedTermsWordsFileLocation := os.Getenv("BANNED_TERMS_WORDS_FILE_LOCATION")
	if bannedTermsWordsFileLocation == "" {
		log.Fatal("❌ BANNED_TERMS_WORDS_FILE_LOCATION not set in environment")
	}

	// Votes
	voteCountToSkip := os.Getenv("VOTE_COUNT_TO_SKIP")
	if voteCountToSkip == "" {
		log.Fatal("❌ VOTE_COUNT_TO_SKIP not set in environment")
	}
	voteCountToSkipNum, err := strconv.Atoi(voteCountToSkip)
	if err != nil {
		log.Fatal("❌ VOTE_COUNT_TO_SKIP not a number")
	}

	// Spotify
	spotifyId := os.Getenv("SPOTIFY_ID")
	if spotifyId == "" {
		log.Fatal("❌ SPOTIFY_ID not set in environment")
	}
	spotifySecret := os.Getenv("SPOTIFY_SECRET")
	if spotifySecret == "" {
		log.Fatal("❌ SPOTIFY_SECRET not set in environment")
	}
	spotifyFallbackPlaylistId := os.Getenv("SPOTIFY_FALLBACK_PLAYLIST_ID")
	if spotifyFallbackPlaylistId == "" {
		log.Fatal("❌ SPOTIFY_FALLBACK_PLAYLIST_ID not set in environment")
	}

	// API
	apiAdminUsername := os.Getenv("API_ADMIN_USERNAME")
	if apiAdminUsername == "" {
		log.Fatal("❌ API_ADMIN_USERNAME not set in environment")
	}
	apiAdminPassword := os.Getenv("API_ADMIN_PASSWORD")
	if apiAdminPassword == "" {
		log.Fatal("❌ API_ADMIN_PASSWORD not set in environment")
	}
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		log.Fatal("❌ API_PORT not set in environment")
	}
	apiAuthCallbackUrl := os.Getenv("API_AUTH_CALLBACK_URL")
	if apiAuthCallbackUrl == "" {
		log.Fatal("❌ API_AUTH_CALLBACK_URL not set in environment")
	}

	return Config{
		DbPath:          dbPath,
		VoteCountToSkip: voteCountToSkipNum,
		Spotify: SpotifyConfig{
			Id:                 spotifyId,
			Secret:             spotifySecret,
			FallbackPlaylistId: spotifyFallbackPlaylistId,
		},
		Api: ApiConfig{
			AdminUsername:   apiAdminUsername,
			AdminPassword:   apiAdminPassword,
			Port:            apiPort,
			AuthCallBackUrl: apiAuthCallbackUrl,
		},
		BannedTerms: BannedTerms{
			WordsFileLocation:  bannedTermsWordsFileLocation,
			TracksFileLocation: bannedTermsTracksFileLocation,
		},
	}
}
