package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"polla/internal/auth"
	"polla/internal/model"
	"polla/internal/ranking"
)

// editLockCutoffSeq returns the chronological seq up to and including which
// matches are temporarily not editable, and whether a lock is configured.
func (s *Server) editLockCutoffSeq(ctx context.Context) (int, bool) {
	lockID, err := s.store.GetEditLockMatchID(ctx)
	if err != nil || lockID == nil {
		return -1, false
	}
	m, err := s.store.GetMatch(ctx, *lockID)
	if err != nil {
		return -1, false
	}
	return m.Seq, true
}

type predictionDTO struct {
	Home          *int   `json:"home"`
	Away          *int   `json:"away"`
	PenaltyWinner string `json:"penaltyWinner"`
}

type matchDTO struct {
	ID           int64          `json:"id"`
	UTCDate      time.Time      `json:"utcDate"`
	Stage        string         `json:"stage"`
	Group        string         `json:"group"`
	HomeTeamName string         `json:"homeTeamName"`
	AwayTeamName string         `json:"awayTeamName"`
	HomeCrest    string         `json:"homeCrest"`
	AwayCrest    string         `json:"awayCrest"`
	Status       string         `json:"status"`
	ScoreHome    *int           `json:"scoreHome"`
	ScoreAway    *int           `json:"scoreAway"`
	Winner       string         `json:"winner"`
	Duration     string         `json:"duration"`
	Knockout     bool           `json:"knockout"`
	Editable     bool           `json:"editable"`
	TempLocked   bool           `json:"tempLocked"`
	Active       bool           `json:"active"`
	Prediction   *predictionDTO `json:"prediction"`
	Points       *int           `json:"points"`
}

type matchesResponse struct {
	Matches       []matchDTO `json:"matches"`
	ActiveMatchID *int64     `json:"activeMatchId"`
	MyRank        int        `json:"myRank"`
	MyPoints      int        `json:"myPoints"`
	TotalCards    int        `json:"totalCards"`
}

// ownedCard loads a card and verifies it belongs to the current player (or that
// the requester is the admin).
func (s *Server) ownedCard(w http.ResponseWriter, r *http.Request, cardID int64) (model.Card, bool) {
	id, _ := auth.IdentityFrom(r.Context())
	card, err := s.store.GetCard(r.Context(), cardID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Cartón no encontrado.")
		return model.Card{}, false
	}
	if !id.IsAdmin && card.PlayerID != id.PlayerID {
		writeError(w, http.StatusForbidden, "Ese cartón no es tuyo.")
		return model.Card{}, false
	}
	return card, true
}

// handleMatches returns all matches with the given card's predictions, the
// editable flag (true until kickoff), and which match is currently active.
func (s *Server) handleMatches(w http.ResponseWriter, r *http.Request) {
	cardID, err := strconv.ParseInt(r.URL.Query().Get("cardId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Falta el cartón.")
		return
	}
	if _, ok := s.ownedCard(w, r, cardID); !ok {
		return
	}
	ctx := r.Context()

	matches, err := s.store.ListMatches(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar los partidos.")
		return
	}
	teams, err := s.store.ListTeams(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar los equipos.")
		return
	}
	cardPreds, err := s.store.ListCardPredictions(ctx, cardID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar los marcadores.")
		return
	}
	predByMatch := map[int64]model.CardPrediction{}
	for _, p := range cardPreds {
		predByMatch[p.MatchID] = p
	}

	now := time.Now()
	lock := time.Duration(s.cfg.LockOffsetMinutes) * time.Minute
	cutoffSeq, locked := s.editLockCutoffSeq(ctx)

	var activeID *int64
	for _, m := range matches {
		if m.UTCDate.After(now) {
			id := m.ID
			activeID = &id
			break
		}
	}
	if activeID == nil && len(matches) > 0 {
		id := matches[len(matches)-1].ID
		activeID = &id
	}

	views := make([]matchDTO, 0, len(matches))
	for _, m := range matches {
		dto := matchDTO{
			ID:           m.ID,
			UTCDate:      m.UTCDate,
			Stage:        m.Stage,
			Group:        m.GroupLetter,
			HomeTeamName: m.HomeTeamName,
			AwayTeamName: m.AwayTeamName,
			Status:       m.Status,
			ScoreHome:    m.ScoreHome,
			ScoreAway:    m.ScoreAway,
			Winner:       m.Winner,
			Duration:     m.Duration,
			Knockout:     !model.IsGroupStage(m.Stage),
			Active:       activeID != nil && *activeID == m.ID,
		}
		dto.TempLocked = locked && m.Seq <= cutoffSeq
		dto.Editable = now.Before(m.UTCDate.Add(-lock)) && !m.Started() && !dto.TempLocked
		if m.HomeTeamID != nil {
			dto.HomeCrest = teams[*m.HomeTeamID].CrestURL
		}
		if m.AwayTeamID != nil {
			dto.AwayCrest = teams[*m.AwayTeamID].CrestURL
		}
		if p, ok := predByMatch[m.ID]; ok {
			dto.Prediction = &predictionDTO{Home: p.Home, Away: p.Away, PenaltyWinner: p.PenaltyWinner}
			if m.Finished() {
				pts := ranking.MatchPoints(m, p)
				dto.Points = &pts
			}
		}
		views = append(views, dto)
	}

	rank, points, total, err := s.ranking.CardRank(ctx, cardID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos calcular el ranking.")
		return
	}

	writeJSON(w, http.StatusOK, matchesResponse{
		Matches:       views,
		ActiveMatchID: activeID,
		MyRank:        rank,
		MyPoints:      points,
		TotalCards:    total,
	})
}

type putPredictionRequest struct {
	Home          *int   `json:"home"`
	Away          *int   `json:"away"`
	PenaltyWinner string `json:"penaltyWinner"`
}

// handlePutPrediction stores or updates a card's marcador, rejecting edits once
// the match has started.
func (s *Server) handlePutPrediction(w http.ResponseWriter, r *http.Request) {
	cardID, err := strconv.ParseInt(r.PathValue("cardId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Cartón inválido.")
		return
	}
	if _, ok := s.ownedCard(w, r, cardID); !ok {
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

	m, err := s.store.GetMatch(r.Context(), matchID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Partido no encontrado.")
		return
	}
	if cutoffSeq, locked := s.editLockCutoffSeq(r.Context()); locked && m.Seq <= cutoffSeq {
		writeError(w, http.StatusConflict, "Este partido está temporalmente no editable.")
		return
	}
	lock := time.Duration(s.cfg.LockOffsetMinutes) * time.Minute
	if m.Started() || !time.Now().Before(m.UTCDate.Add(-lock)) {
		writeError(w, http.StatusConflict, "Este partido ya comenzó; el marcador quedó cerrado.")
		return
	}
	if !validScore(req.Home) || !validScore(req.Away) {
		writeError(w, http.StatusBadRequest, "El marcador debe estar entre 0 y 99.")
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
		s.log.Error("upsert card prediction failed", "err", err)
		writeError(w, http.StatusInternalServerError, "No pudimos guardar tu marcador.")
		return
	}
	writeJSON(w, http.StatusOK, predictionDTO{Home: req.Home, Away: req.Away, PenaltyWinner: penaltyWinner})
}

func validScore(v *int) bool {
	return v == nil || (*v >= 0 && *v <= 99)
}
