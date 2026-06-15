package api

import (
	"net/http"
	"strconv"

	"polla/internal/auth"
)

// handleRanking returns the full per-card ranking and the current user's player
// id so the frontend can highlight their cards.
func (s *Server) handleRanking(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.IdentityFrom(r.Context())
	entries, err := s.ranking.Standings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos calcular el ranking.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ranking":    entries,
		"mePlayerId": id.PlayerID,
	})
}

// handleTeams returns the team ranking (sum of each member's best card) plus
// the players with no team.
func (s *Server) handleTeams(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.IdentityFrom(r.Context())
	teams, sinEquipo, err := s.ranking.TeamStandings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos calcular los equipos.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"teams":     teams,
		"sinEquipo": sinEquipo,
		"mePlayerId": id.PlayerID,
	})
}

// handleTables returns the real-vs-predicted group tables for a given card.
func (s *Server) handleTables(w http.ResponseWriter, r *http.Request) {
	cardID, err := strconv.ParseInt(r.URL.Query().Get("cardId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Falta el cartón.")
		return
	}
	if _, ok := s.ownedCard(w, r, cardID); !ok {
		return
	}
	tables, err := s.ranking.GroupTables(r.Context(), cardID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar las tablas.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"groups": tables})
}
