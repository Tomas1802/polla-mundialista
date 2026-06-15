package api

import (
	"net/http"
	"strconv"
	"time"

	"polla/internal/model"
	"polla/internal/ranking"
)

// scMatch is one row of a card's public scorecard: the official result, the
// player's marcador, and the points/rule it earned (when the match is over).
type scMatch struct {
	ID           int64     `json:"id"`
	UTCDate      time.Time `json:"utcDate"`
	Stage        string    `json:"stage"`
	Group        string    `json:"group"`
	HomeTeamName string    `json:"homeTeamName"`
	AwayTeamName string    `json:"awayTeamName"`
	HomeCrest    string    `json:"homeCrest"`
	AwayCrest    string    `json:"awayCrest"`
	Status       string    `json:"status"`
	ScoreHome    *int      `json:"scoreHome"`
	ScoreAway    *int      `json:"scoreAway"`
	PredHome     *int      `json:"predHome"`
	PredAway     *int      `json:"predAway"`
	Finished     bool      `json:"finished"`
	Points       *int      `json:"points"`
	Rule         string    `json:"rule"`
}

type scorecardResponse struct {
	CardID      int64     `json:"cardId"`
	PlayerName  string    `json:"playerName"`
	CardLabel   string    `json:"cardLabel"`
	TotalPoints int       `json:"totalPoints"`
	Matches     []scMatch `json:"matches"`
}

// handleScorecard returns a card's marcadores with per-match points and the rule
// that produced them, up to the current match. Any authenticated player may view
// any card; only matches that have already kicked off are exposed, so future
// predictions stay private.
func (s *Server) handleScorecard(w http.ResponseWriter, r *http.Request) {
	cardID, err := strconv.ParseInt(r.PathValue("cardId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Cartón inválido.")
		return
	}
	ctx := r.Context()

	card, err := s.store.GetCard(ctx, cardID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Cartón no encontrado.")
		return
	}
	playerName, _ := s.store.GetPlayerName(ctx, card.PlayerID)

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
	preds, err := s.store.ListCardPredictions(ctx, cardID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar los marcadores.")
		return
	}
	predByMatch := map[int64]model.CardPrediction{}
	for _, p := range preds {
		predByMatch[p.MatchID] = p
	}

	rows := make([]scMatch, 0)
	total := 0
	for _, m := range matches {
		if !m.Started() {
			continue // keep future predictions private
		}
		row := scMatch{
			ID:           m.ID,
			UTCDate:      m.UTCDate,
			Stage:        m.Stage,
			Group:        m.GroupLetter,
			HomeTeamName: m.HomeTeamName,
			AwayTeamName: m.AwayTeamName,
			Status:       m.Status,
			ScoreHome:    m.ScoreHome,
			ScoreAway:    m.ScoreAway,
			Finished:     m.Finished(),
		}
		if m.HomeTeamID != nil {
			row.HomeCrest = teams[*m.HomeTeamID].CrestURL
		}
		if m.AwayTeamID != nil {
			row.AwayCrest = teams[*m.AwayTeamID].CrestURL
		}
		p, hasPred := predByMatch[m.ID]
		if hasPred {
			row.PredHome = p.Home
			row.PredAway = p.Away
		}
		if m.Finished() {
			pts, rule := ranking.MatchOutcome(m, p)
			row.Points = &pts
			row.Rule = rule
			total += pts
		}
		rows = append(rows, row)
	}

	writeJSON(w, http.StatusOK, scorecardResponse{
		CardID:      cardID,
		PlayerName:  playerName,
		CardLabel:   card.Label,
		TotalPoints: total,
		Matches:     rows,
	})
}
