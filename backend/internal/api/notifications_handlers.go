package api

import (
	"fmt"
	"net/http"
	"time"

	"polla/internal/model"
	"polla/internal/ranking"
)

type notificationDTO struct {
	ID    string    `json:"id"`
	Icon  string    `json:"icon"`
	Title string    `json:"title"`
	Body  string    `json:"body"`
	TS    time.Time `json:"ts"`
}

// handleNotifications returns the day's pool notifications (results, exact-score
// counts, current podium) with football-flavoured copy. Everything is derived
// on the fly from current state and scoped to today (Colombia), so nothing is
// stored per player — read-state is tracked client-side and resets each day.
func (s *Server) handleNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now()
	start := nowColombiaMidnight(now)
	end := start.Add(24 * time.Hour)

	matches, err := s.store.ListMatches(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar los partidos.")
		return
	}
	preds, err := s.store.ListAllCardPredictions(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar los marcadores.")
		return
	}

	var finishedToday []model.Match
	for _, m := range matches {
		if m.Finished() && !m.UTCDate.Before(start) && m.UTCDate.Before(end) {
			finishedToday = append(finishedToday, m)
		}
	}

	notifs := make([]notificationDTO, 0)

	// Headline: current podium, tied to the latest finished match of the day.
	if len(finishedToday) > 0 {
		last := finishedToday[len(finishedToday)-1]
		if entries, err := s.ranking.Standings(ctx); err == nil && len(entries) > 0 {
			notifs = append(notifs, podiumNotification(last, entries))
		}
	}

	// One "final whistle" notification per match finished today (newest first).
	for i := len(finishedToday) - 1; i >= 0; i-- {
		m := finishedToday[i]
		notifs = append(notifs, finalNotification(m, countExactPredictions(m, preds)))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"date":          start.In(colombiaZone).Format("2006-01-02"),
		"notifications": notifs,
	})
}

func finalNotification(m model.Match, exact int) notificationDTO {
	var body string
	switch exact {
	case 0:
		body = "Nadie clavó el marcador exacto. ¡Qué partido tan bravo! 😅"
	case 1:
		body = "1 cartón clavó el marcador exacto. ¡Olfato de goleador! 🎯"
	default:
		body = fmt.Sprintf("%d cartones clavaron el marcador exacto. 🎯", exact)
	}
	sh, sa := 0, 0
	if m.ScoreHome != nil {
		sh = *m.ScoreHome
	}
	if m.ScoreAway != nil {
		sa = *m.ScoreAway
	}
	return notificationDTO{
		ID:    fmt.Sprintf("final-%d", m.ID),
		Icon:  "⚽",
		Title: fmt.Sprintf("¡Pitazo final! %s %d–%d %s", m.HomeTeamName, sh, sa, m.AwayTeamName),
		Body:  body,
		TS:    m.UTCDate,
	}
}

func podiumNotification(after model.Match, entries []ranking.Entry) notificationDTO {
	parts := []string{"🥇", "🥈", "🥉"}
	body := ""
	for i := 0; i < 3 && i < len(entries); i++ {
		if i > 0 {
			body += "   "
		}
		body += fmt.Sprintf("%s %s (%s) · %d pts", parts[i], entries[i].PlayerName, entries[i].CardLabel, entries[i].Points)
	}
	return notificationDTO{
		ID:    fmt.Sprintf("podium-%d", after.ID),
		Icon:  "🏆",
		Title: "Así está el podio",
		Body:  body,
		TS:    after.UTCDate,
	}
}

// countExactPredictions counts cards whose marcador exactly matched the result.
func countExactPredictions(m model.Match, preds map[int64][]model.CardPrediction) int {
	if m.ScoreHome == nil || m.ScoreAway == nil {
		return 0
	}
	n := 0
	for _, list := range preds {
		for _, p := range list {
			if p.MatchID == m.ID && p.Filled() && *p.Home == *m.ScoreHome && *p.Away == *m.ScoreAway {
				n++
			}
		}
	}
	return n
}
