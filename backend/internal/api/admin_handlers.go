package api

import (
	"net/http"

	"polla/internal/auth"
	"polla/internal/scores"
)

// requireAdmin returns true if the requester is the master admin (else 403).
func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	id, _ := auth.IdentityFrom(r.Context())
	if !id.IsAdmin {
		writeError(w, http.StatusForbidden, "Solo el administrador puede hacer esto.")
		return false
	}
	return true
}

type adminPlayerDTO struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	CardCount     int    `json:"cardCount"`
	HasPin        bool   `json:"hasPin"`        // a PIN has been generated
	MustChangePin bool   `json:"mustChangePin"` // false = player set their own
}

// handleAdminPlayers lists every player with their PIN status (admin only). PIN
// values are never exposed in the app; they live only in the generated CSV.
func (s *Server) handleAdminPlayers(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	players, err := s.store.ListPlayersAdmin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar los jugadores.")
		return
	}
	out := make([]adminPlayerDTO, 0, len(players))
	for _, p := range players {
		out = append(out, adminPlayerDTO{
			ID:            p.ID,
			Name:          p.Name,
			CardCount:     p.CardCount,
			HasPin:        p.HasPin,
			MustChangePin: p.MustChangePin,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"players": out})
}

// handleAdminImportScores imports the cartón CSV files (admin only).
func (s *Server) handleAdminImportScores(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	result, err := scores.Import(r.Context(), s.store, s.cfg.ScoresDir)
	if err != nil {
		s.log.Error("scores import failed", "err", err)
		writeError(w, http.StatusInternalServerError, "No se pudo importar: "+err.Error())
		return
	}
	s.log.Info("scores imported", "players", result.Players, "cards", result.Cards,
		"predictions", result.Predictions)
	writeJSON(w, http.StatusOK, result)
}
