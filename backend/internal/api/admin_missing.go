package api

import (
	"net/http"
	"time"
)

// colombiaZone is UTC-5 (no DST); "today" for the pool is the Colombian day.
var colombiaZone = time.FixedZone("COT", -5*60*60)

type missingCardDTO struct {
	PlayerName string   `json:"playerName"`
	CardLabel  string   `json:"cardLabel"`
	Missing    []string `json:"missing"` // labels of the matches still unfilled
}

type missingReportDTO struct {
	Date       string           `json:"date"`       // dd/mm (Colombia)
	Matches    []string         `json:"matches"`    // today's still-open matches considered
	MatchCount int              `json:"matchCount"` // = len(Matches)
	TotalCards int              `json:"totalCards"` // all cards in the pool
	Cards      []missingCardDTO `json:"cards"`      // cards with >=1 missing marcador
}

// handleAdminMissingToday reports which cards have not filled in today's matches
// that are still open (not yet kicked off), so the admin can nudge the group.
func (s *Server) handleAdminMissingToday(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	ctx := r.Context()

	now := time.Now()
	startCO := nowColombiaMidnight(now)
	endCO := startCO.Add(24 * time.Hour)

	matches, err := s.store.ListMatches(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar los partidos.")
		return
	}
	// Today's matches that are still fillable (kickoff in the future).
	type dayMatch struct {
		id    int64
		label string
	}
	var today []dayMatch
	for _, m := range matches {
		if m.UTCDate.Before(startCO) || !m.UTCDate.Before(endCO) {
			continue
		}
		if !m.UTCDate.After(now) || m.Started() {
			continue // already kicked off — no longer fillable
		}
		today = append(today, dayMatch{id: m.ID, label: m.HomeTeamName + " vs " + m.AwayTeamName})
	}

	cards, err := s.store.ListCardsWithPlayer(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar los cartones.")
		return
	}
	predsByCard, err := s.store.ListAllCardPredictions(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar los marcadores.")
		return
	}

	out := missingReportDTO{
		Date:       startCO.In(colombiaZone).Format("02/01"),
		MatchCount: len(today),
		TotalCards: len(cards),
		Matches:    make([]string, 0, len(today)),
		Cards:      make([]missingCardDTO, 0),
	}
	for _, dm := range today {
		out.Matches = append(out.Matches, dm.label)
	}

	for _, c := range cards {
		filled := map[int64]bool{}
		for _, p := range predsByCard[c.CardID] {
			if p.Filled() {
				filled[p.MatchID] = true
			}
		}
		var missing []string
		for _, dm := range today {
			if !filled[dm.id] {
				missing = append(missing, dm.label)
			}
		}
		if len(missing) > 0 {
			out.Cards = append(out.Cards, missingCardDTO{
				PlayerName: c.PlayerName,
				CardLabel:  c.Label,
				Missing:    missing,
			})
		}
	}

	writeJSON(w, http.StatusOK, out)
}

// nowColombiaMidnight returns the instant of today's 00:00 in Colombia (UTC-5).
func nowColombiaMidnight(now time.Time) time.Time {
	co := now.In(colombiaZone)
	y, mo, d := co.Date()
	return time.Date(y, mo, d, 0, 0, 0, 0, colombiaZone)
}
