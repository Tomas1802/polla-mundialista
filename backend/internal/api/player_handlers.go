package api

import (
	"net/http"

	"polla/internal/auth"
)

// handlePlayers lists all players (with card counts) for the login name picker.
// Public: needed before authenticating.
func (s *Server) handlePlayers(w http.ResponseWriter, r *http.Request) {
	players, err := s.store.ListPlayers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar los jugadores.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"players": players})
}

type cardDTO struct {
	ID          int64  `json:"id"`
	CardNo      int    `json:"cardNo"`
	Label       string `json:"label"`
	Rank        int    `json:"rank"`
	Points      int    `json:"points"`
	ExactScores int    `json:"exactScores"`
}

// handleCards returns the logged-in player's cards with each one's rank/points.
func (s *Server) handleCards(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.IdentityFrom(r.Context())
	if id.PlayerID == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"cards": []cardDTO{}, "totalCards": 0})
		return
	}
	ctx := r.Context()
	cards, err := s.store.ListCardsByPlayer(ctx, id.PlayerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos cargar tus cartones.")
		return
	}
	entries, err := s.ranking.Standings(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No pudimos calcular el ranking.")
		return
	}
	rankByCard := map[int64]int{}
	pointsByCard := map[int64]int{}
	exactByCard := map[int64]int{}
	for _, e := range entries {
		rankByCard[e.CardID] = e.Rank
		pointsByCard[e.CardID] = e.Points
		exactByCard[e.CardID] = e.ExactScores
	}
	out := make([]cardDTO, 0, len(cards))
	for _, c := range cards {
		out = append(out, cardDTO{
			ID:          c.ID,
			CardNo:      c.CardNo,
			Label:       c.Label,
			Rank:        rankByCard[c.ID],
			Points:      pointsByCard[c.ID],
			ExactScores: exactByCard[c.ID],
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"cards": out, "totalCards": len(entries)})
}
