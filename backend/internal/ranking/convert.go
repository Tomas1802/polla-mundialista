// Package ranking turns stored card predictions and official results into points
// and standings, combining the scoring engine (sections 1/3/4 and 2) with the
// standings engine. It is the bridge between persistence and the pure logic.
package ranking

import (
	"polla/internal/model"
	"polla/internal/scoring"
)

// toPrediction adapts a stored card prediction to the scoring engine's type.
func toPrediction(p model.CardPrediction) scoring.Prediction {
	sp := scoring.Prediction{Filled: p.Filled()}
	if p.Home != nil {
		sp.Home = *p.Home
	}
	if p.Away != nil {
		sp.Away = *p.Away
	}
	switch p.PenaltyWinner {
	case "HOME":
		sp.PenaltyWinner = scoring.SideHome
	case "AWAY":
		sp.PenaltyWinner = scoring.SideAway
	}
	return sp
}

// toResult adapts an official match result to the scoring engine's type.
func toResult(m model.Match) scoring.Result {
	r := scoring.Result{}
	if m.ScoreHome != nil {
		r.Home = *m.ScoreHome
	}
	if m.ScoreAway != nil {
		r.Away = *m.ScoreAway
	}
	switch m.Winner {
	case "HOME_WIN":
		r.Winner = scoring.SideHome
	case "AWAY_WIN":
		r.Winner = scoring.SideAway
	default:
		r.Winner = scoring.SideNone
	}
	return r
}

// MatchPoints scores a single match for a card prediction (0 if not finished).
func MatchPoints(m model.Match, p model.CardPrediction) int {
	if !m.Finished() {
		return 0
	}
	return matchPoints(m, p)
}

func matchPoints(m model.Match, p model.CardPrediction) int {
	res := toResult(m)
	pred := toPrediction(p)
	switch {
	case model.IsGroupStage(m.Stage):
		return scoring.ScoreGroupMatch(pred, res)
	case model.IsFinal(m.Stage):
		return scoring.ScoreKnockoutMatch(pred, res, scoring.FinalTiers)
	default:
		return scoring.ScoreKnockoutMatch(pred, res, scoring.KnockoutTiers)
	}
}

// MatchOutcome scores a finished match for a card and returns the points plus a
// short Spanish label of the rule that produced them. Returns (0, "") if the
// match is not finished.
func MatchOutcome(m model.Match, p model.CardPrediction) (int, string) {
	if !m.Finished() {
		return 0, ""
	}
	res := toResult(m)
	pred := toPrediction(p)
	var o scoring.Outcome
	switch {
	case model.IsGroupStage(m.Stage):
		o = scoring.ScoreGroupMatchExplained(pred, res)
	case model.IsFinal(m.Stage):
		o = scoring.ScoreKnockoutMatchExplained(pred, res, scoring.FinalTiers)
	default:
		o = scoring.ScoreKnockoutMatchExplained(pred, res, scoring.KnockoutTiers)
	}
	return o.Points, ruleLabel(o.Rule)
}

// ruleLabel maps a scoring rule to a short Spanish description for the UI.
func ruleLabel(r scoring.Rule) string {
	switch r {
	case scoring.RuleExact:
		return "Marcador exacto"
	case scoring.RuleReversed:
		return "Exacto pero al revés"
	case scoring.RuleWinner:
		return "Ganador acertado"
	case scoring.RuleDraw:
		return "Empate acertado"
	case scoring.RuleExactAdv:
		return "Exacto y clasificado"
	case scoring.RuleExactNoAdv:
		return "Exacto, otro clasificó"
	case scoring.RuleAdv:
		return "Clasificado acertado"
	case scoring.RuleWrong:
		return "Resultado incorrecto"
	default:
		return "Sin pronóstico"
	}
}

// isExactScore reports whether a prediction nailed the exact marcador (the
// ranking tie-breaker).
func isExactScore(m model.Match, p model.CardPrediction) bool {
	if !p.Filled() || !m.Finished() || m.ScoreHome == nil || m.ScoreAway == nil {
		return false
	}
	return *p.Home == *m.ScoreHome && *p.Away == *m.ScoreAway
}

// indexByMatch builds a match-id → prediction lookup.
func indexByMatch(preds []model.CardPrediction) map[int64]model.CardPrediction {
	out := make(map[int64]model.CardPrediction, len(preds))
	for _, p := range preds {
		out[p.MatchID] = p
	}
	return out
}
