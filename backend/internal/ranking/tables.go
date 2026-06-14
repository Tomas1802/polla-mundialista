package ranking

import (
	"context"
	"sort"

	"polla/internal/model"
	"polla/internal/scoring"
	"polla/internal/standings"
)

// TableRow is one row of a group table, ready for display.
type TableRow struct {
	TeamID       int    `json:"teamId"`
	TeamName     string `json:"teamName"`
	Played       int    `json:"played"`
	Won          int    `json:"won"`
	Drawn        int    `json:"drawn"`
	Lost         int    `json:"lost"`
	GoalsFor     int    `json:"goalsFor"`
	GoalsAgainst int    `json:"goalsAgainst"`
	GoalDiff     int    `json:"goalDiff"`
	Points       int    `json:"points"`
}

// GroupTable compares a user's predicted standings with the real ones for one
// group, and the section-2 points earned once the group is decided.
type GroupTable struct {
	Group      string     `json:"group"`
	Finished   bool       `json:"finished"`
	UserPoints int        `json:"userPoints"`
	Real       []TableRow `json:"real"`
	Predicted  []TableRow `json:"predicted"`
}

// GroupTables builds, for the "Tablas" tab, the real-vs-predicted comparison of
// every group for the given card, sorted by group letter.
func (s *Service) GroupTables(ctx context.Context, cardID int64) ([]GroupTable, error) {
	matches, err := s.store.ListMatches(ctx)
	if err != nil {
		return nil, err
	}
	teams, err := s.store.ListTeams(ctx)
	if err != nil {
		return nil, err
	}
	cardPreds, err := s.store.ListCardPredictions(ctx, cardID)
	if err != nil {
		return nil, err
	}
	preds := indexByMatch(cardPreds)

	teamIDsByGroup := groupTeamIDs(teams)
	letters := make([]string, 0, len(teamIDsByGroup))
	for letter := range teamIDsByGroup {
		letters = append(letters, letter)
	}
	sort.Strings(letters)

	out := make([]GroupTable, 0, len(letters))
	for _, letter := range letters {
		ids := teamIDsByGroup[letter]
		gms := groupMatches(letter, matches)

		realRows, finished := realStandings(ids, gms)
		predRows := predictedStandings(ids, gms, preds)

		gt := GroupTable{
			Group:     letter,
			Finished:  finished,
			Real:      toTableRows(realRows, teams),
			Predicted: toTableRows(predRows, teams),
		}
		if finished && len(ids) == 4 && hasGroupPredictions(gms, preds) {
			gt.UserPoints = scoring.ScoreGroupPositions(toGroupOrder(predRows), toGroupOrder(realRows))
		}
		out = append(out, gt)
	}
	return out, nil
}

func toTableRows(rows []standings.Row, teams map[int64]model.Team) []TableRow {
	out := make([]TableRow, len(rows))
	for i, r := range rows {
		out[i] = TableRow{
			TeamID:       r.TeamID,
			TeamName:     teams[int64(r.TeamID)].Name,
			Played:       r.Played,
			Won:          r.Won,
			Drawn:        r.Drawn,
			Lost:         r.Lost,
			GoalsFor:     r.GoalsFor,
			GoalsAgainst: r.GoalsAgainst,
			GoalDiff:     r.GoalDiff(),
			Points:       r.Points,
		}
	}
	return out
}
