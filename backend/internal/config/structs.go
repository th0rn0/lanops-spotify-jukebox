package config

type Config struct {
	DbPath          string
	VoteCountToSkip int
	Spotify         SpotifyConfig
	Api             ApiConfig
	BannedTerms     BannedTerms
}

type SpotifyConfig struct {
	Id                 string
	Secret             string
	FallbackPlaylistId string
}

type ApiConfig struct {
	AdminUsername   string
	AdminPassword   string
	Port            string
	AuthCallBackUrl string
}

type BannedTerms struct {
	WordsFileLocation  string
	TracksFileLocation string
}
