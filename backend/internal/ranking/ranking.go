package ranking

import (
	"context"
	"sort"

	"polla/internal/db"
	"polla/internal/model"
	"polla/internal/scoring"
)

// Entry is one card's place in the ranking.
type Entry struct {
	CardID      int64  `json:"cardId"`
	PlayerID    int64  `json:"playerId"`
	PlayerName  string `json:"playerName"`
	CardNo      int    `json:"cardNo"`
	CardLabel   string `json:"cardLabel"`
	Points      int    `json:"points"`
	ExactScores int    `json:"exactScores"`
	Rank        int    `json:"rank"`
}

// Service computes points and standings from stored data.
type Service struct {
	store *db.DB
}

// New builds a ranking service.
func New(store *db.DB) *Service { return &Service{store: store} }

// Standings computes the full per-card ranking, ordered best-first, with shared
// ranks broken by the number of exact marcadores.
func (s *Service) Standings(ctx context.Context) ([]Entry, error) {
	cards, err := s.store.ListAllCards(ctx)
	if err != nil {
		return nil, err
	}
	players, err := s.store.ListPlayers(ctx)
	if err != nil {
		return nil, err
	}
	matches, err := s.store.ListMatches(ctx)
	if err != nil {
		return nil, err
	}
	teams, err := s.store.ListTeams(ctx)
	if err != nil {
		return nil, err
	}
	predsByCard, err := s.store.ListAllCardPredictions(ctx)
	if err != nil {
		return nil, err
	}

	playerName := map[int64]string{}
	for _, p := range players {
		playerName[p.ID] = p.Name
	}

	type groupInfo struct {
		teamIDs   []int
		gms       []model.Match
		realOrder scoring.GroupOrder
		finished  bool
	}
	groups := map[string]groupInfo{}
	for letter, ids := range groupTeamIDs(teams) {
		gms := groupMatches(letter, matches)
		rows, finished := realStandings(ids, gms)
		groups[letter] = groupInfo{teamIDs: ids, gms: gms, realOrder: toGroupOrder(rows), finished: finished}
	}

	entries := make([]Entry, 0, len(cards))
	for _, card := range cards {
		preds := indexByMatch(predsByCard[card.ID])

		points, exacts := 0, 0
		for _, m := range matches {
			if !m.Finished() {
				continue
			}
			p := preds[m.ID]
			points += matchPoints(m, p)
			if isExactScore(m, p) {
				exacts++
			}
		}
		for _, gi := range groups {
			if !gi.finished || len(gi.teamIDs) != 4 || !hasGroupPredictions(gi.gms, preds) {
				continue
			}
			predOrder := toGroupOrder(predictedStandings(gi.teamIDs, gi.gms, preds))
			points += scoring.ScoreGroupPositions(predOrder, gi.realOrder)
		}

		entries = append(entries, Entry{
			CardID:      card.ID,
			PlayerID:    card.PlayerID,
			PlayerName:  playerName[card.PlayerID],
			CardNo:      card.CardNo,
			CardLabel:   card.Label,
			Points:      points,
			ExactScores: exacts,
		})
	}

	assignRanks(entries)
	return entries, nil
}

// assignRanks sorts entries and assigns shared ranks (1, 2, 2, 4, ...).
func assignRanks(entries []Entry) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Points != entries[j].Points {
			return entries[i].Points > entries[j].Points
		}
		if entries[i].ExactScores != entries[j].ExactScores {
			return entries[i].ExactScores > entries[j].ExactScores
		}
		if entries[i].PlayerName != entries[j].PlayerName {
			return entries[i].PlayerName < entries[j].PlayerName
		}
		return entries[i].CardNo < entries[j].CardNo
	})
	// Dense ranking: tied entries share a place and the next distinct score
	// continues consecutively (1, 2, 2, 3, ...) — no skipped numbers.
	rank := 0
	for i := range entries {
		if i == 0 || entries[i].Points != entries[i-1].Points || entries[i].ExactScores != entries[i-1].ExactScores {
			rank++
		}
		entries[i].Rank = rank
	}
}

// CardRank returns the rank and points of a single card from the standings.
func (s *Service) CardRank(ctx context.Context, cardID int64) (rank, points, total int, err error) {
	entries, err := s.Standings(ctx)
	if err != nil {
		return 0, 0, 0, err
	}
	for _, e := range entries {
		if e.CardID == cardID {
			return e.Rank, e.Points, len(entries), nil
		}
	}
	return 0, 0, len(entries), nil
}
