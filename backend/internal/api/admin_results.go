package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type adminResultMatchDTO struct {
	ID           int64     `json:"id"`
	UTCDate      time.Time `json:"utcDate"`
	Stage        string    `json:"stage"`
	Group        string    `json:"group"`
	Home         string    `json:"home"`
	Away         string    `json:"away"`
	Status       string    `json:"status"`
	ScoreHome    *int      `json:"scoreHome"`
	ScoreAway    *int      `json:"scoreAway"`
	ResultManual bool      `json:"resultManual"`
	Started      bool      `json:"started"`
	Finished     bool      `json:"finished"`
}

// handleAdminMatches lists matches that have already started/finished (most
// recent first) so the admin can correct a wrong official result.
func (s *Server) handleAdminMatches(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	rows, err := s.store.ListMatchResults(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar los partidos.")
		return
	}
	out := make([]adminResultMatchDTO, 0)
	for i := len(rows) - 1; i >= 0; i-- { // reverse chronological
		m := rows[i]
		started := matchStarted(m.Status)
		if !started {
			continue // only played matches can have a result to correct
		}
		out = append(out, adminResultMatchDTO{
			ID: m.ID, UTCDate: m.UTCDate, Stage: m.Stage, Group: m.GroupLetter,
			Home: m.HomeTeamName, Away: m.AwayTeamName, Status: m.Status,
			ScoreHome: m.ScoreHome, ScoreAway: m.ScoreAway, ResultManual: m.ResultManual,
			Started: started, Finished: m.Status == "FINISHED",
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"matches": out})
}

type adminResultRequest struct {
	Home *int `json:"home"`
	Away *int `json:"away"`
}

// handleAdminSetResult overrides a match's official result (marks it manual so
// the sync won't revert it). Points recompute automatically from this result.
func (s *Server) handleAdminSetResult(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	matchID, err := strconv.ParseInt(r.PathValue("matchId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Partido inválido.")
		return
	}
	var req adminResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Solicitud inválida.")
		return
	}
	if req.Home == nil || req.Away == nil || !validScore(req.Home) || !validScore(req.Away) {
		writeError(w, http.StatusBadRequest, "El marcador debe estar entre 0 y 99.")
		return
	}
	winner := "DRAW"
	if *req.Home > *req.Away {
		winner = "HOME_WIN"
	} else if *req.Away > *req.Home {
		winner = "AWAY_WIN"
	}
	if err := s.store.SetMatchResult(r.Context(), matchID, *req.Home, *req.Away, winner); err != nil {
		s.log.Error("admin set result failed", "err", err)
		writeError(w, http.StatusInternalServerError, "No pudimos guardar el resultado.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleAdminConfirmResult freezes a match's current (service) result so the
// sync stops consulting/overwriting it — without changing the score.
func (s *Server) handleAdminConfirmResult(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	matchID, err := strconv.ParseInt(r.PathValue("matchId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Partido inválido.")
		return
	}
	m, err := s.store.GetMatch(r.Context(), matchID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Partido no encontrado.")
		return
	}
	if !m.Finished() || m.ScoreHome == nil || m.ScoreAway == nil {
		writeError(w, http.StatusConflict, "El partido aún no tiene un resultado para confirmar.")
		return
	}
	winner := m.Winner
	if winner == "" {
		winner = "DRAW"
		if *m.ScoreHome > *m.ScoreAway {
			winner = "HOME_WIN"
		} else if *m.ScoreAway > *m.ScoreHome {
			winner = "AWAY_WIN"
		}
	}
	if err := s.store.SetMatchResult(r.Context(), matchID, *m.ScoreHome, *m.ScoreAway, winner); err != nil {
		s.log.Error("admin confirm result failed", "err", err)
		writeError(w, http.StatusInternalServerError, "No pudimos confirmar el resultado.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleAdminClearResult re-enables automatic syncing for a match.
func (s *Server) handleAdminClearResult(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	matchID, err := strconv.ParseInt(r.PathValue("matchId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Partido inválido.")
		return
	}
	if err := s.store.ClearMatchManual(r.Context(), matchID); err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos revertir el resultado.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// matchStarted mirrors model.Match.Started for a bare status string.
func matchStarted(status string) bool {
	switch status {
	case "IN_PLAY", "PAUSED", "FINISHED", "SUSPENDED", "AWARDED":
		return true
	default:
		return false
	}
}
