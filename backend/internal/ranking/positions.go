package ranking

import (
	"polla/internal/model"
	"polla/internal/scoring"
	"polla/internal/standings"
)

// groupTeamIDs maps each group letter to its four team ids.
func groupTeamIDs(teams map[int64]model.Team) map[string][]int {
	out := map[string][]int{}
	for _, t := range teams {
		if t.GroupLetter != "" {
			out[t.GroupLetter] = append(out[t.GroupLetter], int(t.ID))
		}
	}
	return out
}

// groupMatches returns the group-stage matches of a given group.
func groupMatches(letter string, matches []model.Match) []model.Match {
	var out []model.Match
	for _, m := range matches {
		if model.IsGroupStage(m.Stage) && m.GroupLetter == letter {
			out = append(out, m)
		}
	}
	return out
}

// realStandings builds the official table for a group from actual results,
// counting only finished matches. allFinished reports whether the whole group
// is decided (required before section-2 points are awarded).
func realStandings(teamIDs []int, gms []model.Match) (rows []standings.Row, allFinished bool) {
	sms := make([]standings.Match, 0, len(gms))
	allFinished = len(gms) > 0
	for _, m := range gms {
		if m.HomeTeamID == nil || m.AwayTeamID == nil {
			allFinished = false
			continue
		}
		played := m.Finished()
		if !played {
			allFinished = false
		}
		hg, ag := 0, 0
		if played {
			hg, ag = *m.ScoreHome, *m.ScoreAway
		}
		sms = append(sms, standings.Match{
			HomeTeamID: int(*m.HomeTeamID), AwayTeamID: int(*m.AwayTeamID),
			HomeGoals: hg, AwayGoals: ag, Played: played,
		})
	}
	return standings.Compute(teamIDs, sms), allFinished
}

// predictedStandings builds a card's predicted table for a group from its
// marcadores (unfilled predictions count as unplayed).
func predictedStandings(teamIDs []int, gms []model.Match, preds map[int64]model.CardPrediction) []standings.Row {
	sms := make([]standings.Match, 0, len(gms))
	for _, m := range gms {
		if m.HomeTeamID == nil || m.AwayTeamID == nil {
			continue
		}
		p, ok := preds[m.ID]
		played := ok && p.Filled()
		hg, ag := 0, 0
		if played {
			hg, ag = *p.Home, *p.Away
		}
		sms = append(sms, standings.Match{
			HomeTeamID: int(*m.HomeTeamID), AwayTeamID: int(*m.AwayTeamID),
			HomeGoals: hg, AwayGoals: ag, Played: played,
		})
	}
	return standings.Compute(teamIDs, sms)
}

// hasGroupPredictions reports whether the card filled at least one marcador for
// the group. With no predictions at all, section-2 points are 0 (an empty form
// scores nothing), not the "no acierta nada" floor.
func hasGroupPredictions(gms []model.Match, preds map[int64]model.CardPrediction) bool {
	for _, m := range gms {
		if p, ok := preds[m.ID]; ok && p.Filled() {
			return true
		}
	}
	return false
}

// toGroupOrder takes the first four team ids of an ordered table.
func toGroupOrder(rows []standings.Row) scoring.GroupOrder {
	var g scoring.GroupOrder
	for i := 0; i < 4 && i < len(rows); i++ {
		g[i] = rows[i].TeamID
	}
	return g
}
