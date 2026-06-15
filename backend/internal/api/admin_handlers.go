package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"polla/internal/auth"
	"polla/internal/model"
	"polla/internal/scores"
)

type settingsMatchDTO struct {
	ID      int64     `json:"id"`
	Seq     int       `json:"seq"`
	Home    string    `json:"home"`
	Away    string    `json:"away"`
	UTCDate time.Time `json:"utcDate"`
	Group   string    `json:"group"`
}

// handleAdminGetSettings returns the current edit-lock cutoff and the list of
// matches (chronological) for the admin to pick from.
func (s *Server) handleAdminGetSettings(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	lockID, err := s.store.GetEditLockMatchID(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar los ajustes.")
		return
	}
	matches, err := s.store.ListMatches(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar los partidos.")
		return
	}
	out := make([]settingsMatchDTO, 0, len(matches))
	for _, m := range matches {
		out = append(out, settingsMatchDTO{
			ID: m.ID, Seq: m.Seq, Home: m.HomeTeamName, Away: m.AwayTeamName,
			UTCDate: m.UTCDate, Group: m.GroupLetter,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"editLockUntilMatchId": lockID,
		"matches":              out,
	})
}

type adminSettingsRequest struct {
	EditLockUntilMatchID *int64 `json:"editLockUntilMatchId"`
}

// handleAdminSetSettings sets (or clears) the edit-lock cutoff match.
func (s *Server) handleAdminSetSettings(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	var req adminSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Solicitud inválida.")
		return
	}
	if req.EditLockUntilMatchID != nil {
		if _, err := s.store.GetMatch(r.Context(), *req.EditLockUntilMatchID); err != nil {
			writeError(w, http.StatusBadRequest, "Partido inválido.")
			return
		}
	}
	if err := s.store.SetEditLockMatchID(r.Context(), req.EditLockUntilMatchID); err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos guardar el ajuste.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

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

// handleAdminCards lists every card (player + label) for the correction picker.
func (s *Server) handleAdminCards(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	cards, err := s.store.ListCardsWithPlayer(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar los cartones.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"cards": cards})
}

// handleAdminEditPrediction lets the admin correct any card's marcador,
// including already-played matches (no kickoff lock). Admin only.
func (s *Server) handleAdminEditPrediction(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	cardID, err := strconv.ParseInt(r.PathValue("cardId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Cartón inválido.")
		return
	}
	matchID, err := strconv.ParseInt(r.PathValue("matchId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Partido inválido.")
		return
	}
	var req putPredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Solicitud inválida.")
		return
	}
	if !validScore(req.Home) || !validScore(req.Away) {
		writeError(w, http.StatusBadRequest, "El marcador debe estar entre 0 y 99.")
		return
	}
	m, err := s.store.GetMatch(r.Context(), matchID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Partido no encontrado.")
		return
	}
	penaltyWinner := ""
	if !model.IsGroupStage(m.Stage) && (req.PenaltyWinner == "HOME" || req.PenaltyWinner == "AWAY") {
		penaltyWinner = req.PenaltyWinner
	}
	if err := s.store.UpsertCardPrediction(r.Context(), model.CardPrediction{
		CardID:        cardID,
		MatchID:       matchID,
		Home:          req.Home,
		Away:          req.Away,
		PenaltyWinner: penaltyWinner,
	}); err != nil {
		s.log.Error("admin edit prediction failed", "err", err)
		writeError(w, http.StatusInternalServerError, "No pudimos guardar el marcador.")
		return
	}
	writeJSON(w, http.StatusOK, predictionDTO{Home: req.Home, Away: req.Away, PenaltyWinner: penaltyWinner})
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
