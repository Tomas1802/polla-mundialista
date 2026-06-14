// Package api exposes the REST HTTP layer: routing, CORS, authentication, and
// the handlers for matches, predictions, ranking and the group tables.
package api

import (
	"log/slog"
	"net/http"

	"polla/internal/auth"
	"polla/internal/config"
	"polla/internal/db"
	"polla/internal/ranking"
)

// Server holds the handler dependencies.
type Server struct {
	cfg      config.Config
	store    *db.DB
	sessions *auth.Sessions
	authn    *auth.Authenticator
	ranking  *ranking.Service
	log      *slog.Logger
}

// NewServer wires the API dependencies.
func NewServer(cfg config.Config, store *db.DB, sessions *auth.Sessions, log *slog.Logger) *Server {
	return &Server{
		cfg:      cfg,
		store:    store,
		sessions: sessions,
		authn:    auth.NewAuthenticator(sessions, store),
		ranking:  ranking.New(store),
		log:      log,
	}
}

// Handler builds the routed, CORS-wrapped HTTP handler.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Public endpoints.
	mux.HandleFunc("GET /api/players", s.handlePlayers) // for the login name picker
	mux.HandleFunc("POST /api/auth/login", s.handleLogin)
	mux.HandleFunc("POST /api/auth/admin-login", s.handleAdminLogin)
	mux.HandleFunc("POST /api/auth/logout", s.handleLogout)

	// Protected endpoints (mounted under /api/ behind the auth middleware).
	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/me", s.handleMe)
	protected.HandleFunc("POST /api/auth/change-pin", s.handleChangePin)
	protected.HandleFunc("GET /api/cards", s.handleCards)
	protected.HandleFunc("GET /api/matches", s.handleMatches)
	protected.HandleFunc("PUT /api/cards/{cardId}/predictions/{matchId}", s.handlePutPrediction)
	protected.HandleFunc("GET /api/ranking", s.handleRanking)
	protected.HandleFunc("GET /api/tables", s.handleTables)
	protected.HandleFunc("GET /api/admin/players", s.handleAdminPlayers)
	protected.HandleFunc("POST /api/admin/import-scores", s.handleAdminImportScores)
	mux.Handle("/api/", s.authn.Middleware(protected))

	return s.cors(mux)
}

// cors applies the CORS policy. Credentials are allowed, so the origin must be
// the exact configured frontend origin.
func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		if r.Header.Get("Origin") == s.cfg.AllowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", s.cfg.AllowedOrigin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) setSessionCookie(w http.ResponseWriter, token string) {
	c := &http.Cookie{
		Name:     auth.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(s.sessions.TTL().Seconds()),
	}
	if s.cfg.CookieSecure {
		c.Secure = true
		c.SameSite = http.SameSiteNoneMode
	} else {
		c.SameSite = http.SameSiteLaxMode
	}
	http.SetCookie(w, c)
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	c := &http.Cookie{
		Name:     auth.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	if s.cfg.CookieSecure {
		c.Secure = true
		c.SameSite = http.SameSiteNoneMode
	} else {
		c.SameSite = http.SameSiteLaxMode
	}
	http.SetCookie(w, c)
}
