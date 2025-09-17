package api

import (
	"lanops/spotify-jukebox/api/handlers"
	"lanops/spotify-jukebox/internal/config"
	"lanops/spotify-jukebox/internal/jukebox"
	"os"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func SetupRouter(cfg config.Config, c *jukebox.Client, logger zerolog.Logger) *gin.Engine {
	logger.Info().Msg("Loading API")
	gin.DefaultWriter = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", status).
			Dur("latency", latency).
			Msg("request handled")
	})
	r.Use(cors.Default())
	// Handlers
	handlers := handlers.New(c)
	authorized := r.Group("", gin.BasicAuth(gin.Accounts{
		cfg.Api.AdminUsername: cfg.Api.AdminPassword,
	}))

	// Set Rate Limiting
	rateLimitVoteToSkipMiddleWareStore := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Minute * 4,
		Limit: 1,
	})

	rateLimitVoteToSkipMiddleWare := ratelimit.RateLimiter(rateLimitVoteToSkipMiddleWareStore, &ratelimit.Options{
		ErrorHandler: func(c *gin.Context, info ratelimit.Info) {
			c.JSON(429, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
		},
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP() + c.Request.UserAgent()
		},
	})

	// Voting
	r.POST("/votes/skip", rateLimitVoteToSkipMiddleWare, handlers.VoteToSkip)
	// Tracks
	r.GET("/tracks", handlers.GetTracks)
	r.POST("/tracks/add", handlers.AddTrack)
	r.GET("/tracks/current", handlers.GetCurrentTrack)
	r.GET("/tracks/:trackId", handlers.GetTrackById)
	// Search
	r.GET("/search/:searchTerm", handlers.Search)
	// Auth
	r.GET("/auth/callback", handlers.AuthCallback)
	authorized.GET("/auth/login", handlers.AuthLogin)
	// Player
	authorized.POST("/player/start", handlers.PlayerStart)
	authorized.POST("/player/stop", handlers.PlayerStop)
	authorized.POST("/player/volume", handlers.PlayerSetVolume)
	authorized.GET("/player/volume", handlers.PlayerGetVolume)
	authorized.POST("/player/skip", handlers.PlayerSkip)
	authorized.POST("/player/pause", handlers.PlayerPause)

	return r
}
