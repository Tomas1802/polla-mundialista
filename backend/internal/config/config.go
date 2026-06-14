// Package config loads runtime configuration from environment variables. All
// secrets (database URL, football-data token, JWT secret, admin PIN) come from
// the environment so nothing sensitive lives in the repo.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Port the HTTP server listens on. Cloud Run sets $PORT.
	Port string
	// DatabaseURL is a PostgreSQL connection string (Cloud SQL or local).
	DatabaseURL string
	// FootballDataToken is the X-Auth-Token for football-data.org.
	FootballDataToken string
	// CompetitionCode for the tournament (World Cup = "WC").
	CompetitionCode string
	// JWTSecret signs our own session tokens.
	JWTSecret string
	// AdminPin is the master PIN that grants admin access (import, PIN list).
	AdminPin string
	// AllowedOrigin is the frontend origin allowed by CORS.
	AllowedOrigin string
	// SessionTTL is how long a session JWT stays valid (sessions persist until
	// logout, so this is intentionally long).
	SessionTTL time.Duration
	// LockOffsetMinutes is how many minutes before kickoff a marcador locks.
	LockOffsetMinutes int
	// CookieSecure marks the session cookie Secure + SameSite=None (production).
	CookieSecure bool
	// SyncIntervalMinutes is how often the football-data sync policy is checked.
	SyncIntervalMinutes int
	// ScoresDir is the folder with the cartón CSV files for the import.
	ScoresDir string
}

// Load reads configuration from the environment, applying sensible defaults for
// local development and returning an error if a required secret is missing.
func Load() (Config, error) {
	// Convenience for local dev: load a .env file if one exists. Real
	// environment variables always take precedence.
	loadDotEnv(".env")

	c := Config{
		Port:                getenv("PORT", "8080"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		FootballDataToken:   os.Getenv("FOOTBALL_DATA_TOKEN"),
		CompetitionCode:     getenv("COMPETITION_CODE", "WC"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		AdminPin:            os.Getenv("ADMIN_PIN"),
		AllowedOrigin:       getenv("ALLOWED_ORIGIN", "http://localhost:5173"),
		SessionTTL:          getdur("SESSION_TTL", 365*24*time.Hour),
		LockOffsetMinutes:   getint("LOCK_OFFSET_MINUTES", 0),
		CookieSecure:        getbool("COOKIE_SECURE", false),
		SyncIntervalMinutes: getint("SYNC_INTERVAL_MINUTES", 10),
		ScoresDir:           getenv("SCORES_DIR", "../scores"),
	}

	var missing []string
	if c.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %v", missing)
	}
	return c, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getint(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getbool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getdur(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
