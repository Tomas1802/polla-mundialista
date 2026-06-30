package scoring

// Rule is a UI-neutral code identifying which scoring rule produced a match's
// points. The presentation layer maps it to human text.
type Rule string

const (
	RuleNone       Rule = "none"        // no/half prediction → 0
	RuleExact      Rule = "exact"       // exact score
	RuleReversed   Rule = "reversed"    // exact score, teams reversed (group)
	RuleWinner     Rule = "winner"      // right winner, wrong score (group)
	RuleDraw       Rule = "draw"        // right draw, wrong score (group)
	RuleWrong      Rule = "wrong"       // wrong result
	RuleExactAdv   Rule = "exact_adv"   // exact score and right team advancing (knockout)
	RuleExactNoAdv Rule = "exact_noadv" // exact score, wrong team advanced (knockout)
	RuleAdv        Rule = "adv"         // right team advancing, wrong score (knockout)
)

// Outcome is the result of scoring a single match: the points awarded and the
// rule that produced them.
type Outcome struct {
	Points int
	Rule   Rule
}

// ScoreGroupMatch implements section 1 of the reglamento for group-stage
// matches, where the two teams are fixed by the fixture, so only the marcador
// is predicted:
//
//	7  exact score (winner/draw orientation respected)
//	5  right score but with the teams reversed (only when there is a winner)
//	4  right winner, wrong score
//	3  correctly predicted a draw, wrong score
//	1  predicted, but wrong result direction
//	0  empty or half-filled
func ScoreGroupMatch(p Prediction, r Result) int {
	return ScoreGroupMatchExplained(p, r).Points
}

// ScoreGroupMatchExplained scores a group match and also reports the rule used.
func ScoreGroupMatchExplained(p Prediction, r Result) Outcome {
	if !p.Filled {
		return Outcome{0, RuleNone}
	}
	if p.Home == r.Home && p.Away == r.Away {
		return Outcome{7, RuleExact}
	}
	// "Acierta el marcador, pero con los equipos al revés" — e.g. predicted 2-1
	// when the result was 1-2. Only meaningful when there is a winner; a draw
	// reversed is identical to the exact score and is already handled above.
	if p.Home == r.Away && p.Away == r.Home && !r.isDraw() {
		return Outcome{5, RuleReversed}
	}
	predDraw := p.Home == p.Away
	if !r.isDraw() && !predDraw && sign(p.Home-p.Away) == sign(r.Home-r.Away) {
		return Outcome{4, RuleWinner}
	}
	if r.isDraw() && predDraw {
		return Outcome{3, RuleDraw}
	}
	return Outcome{1, RuleWrong}
}

// KnockoutPoints holds the four scoring tiers of a knockout-style match. It is
// a value (not hard-coded constants) so the organizer can tune the system
// without touching the scoring logic.
type KnockoutPoints struct {
	ExactAndWinner int // right score AND right team advancing
	ExactNotWinner int // right score, wrong team advancing (decided on penalties)
	WinnerNotExact int // right team advancing, wrong score
	Nothing        int // predicted, but nothing right
}

// Default tiers from the reglamento. KnockoutTiers covers dieciseisavos,
// octavos, cuartos, semifinales and (per the organizer) the third-place match.
// FinalTiers covers the final.
var (
	KnockoutTiers = KnockoutPoints{ExactAndWinner: 7, ExactNotWinner: 5, WinnerNotExact: 3, Nothing: 1}
	FinalTiers    = KnockoutPoints{ExactAndWinner: 10, ExactNotWinner: 8, WinnerNotExact: 6, Nothing: 1}
)

// ScoreKnockoutMatch implements sections 3 and 4. The marcador is evaluated on
// the full-time score with extra time included; penalties only decide which
// team advances (Result.Winner). An empty/half-filled prediction scores 0.
//
// Note that "exact score but wrong winner" can only occur on a draw scoreline
// (e.g. predicted 1-1 and picked the wrong team to win the shootout), because
// for any non-draw score the winner is implied by the score itself.
func ScoreKnockoutMatch(p Prediction, r Result, t KnockoutPoints) int {
	return ScoreKnockoutMatchExplained(p, r, t).Points
}

// ScoreKnockoutMatchExplained scores a knockout match and reports the rule used.
func ScoreKnockoutMatchExplained(p Prediction, r Result, t KnockoutPoints) Outcome {
	if !p.Filled {
		return Outcome{0, RuleNone}
	}
	exact := p.Home == r.Home && p.Away == r.Away
	// Same goals, teams swapped (predicted 1-2, result 2-1): the marcador is
	// right but the winner is wrong, so it scores the same tier as an exact
	// score with the wrong team advancing. A reversed draw equals the exact
	// score and is already caught above.
	reversed := !r.isDraw() && p.Home == r.Away && p.Away == r.Home
	correctWinner := r.Winner != SideNone && p.winner() == r.Winner
	switch {
	case exact && correctWinner:
		return Outcome{t.ExactAndWinner, RuleExactAdv}
	case exact && !correctWinner:
		return Outcome{t.ExactNotWinner, RuleExactNoAdv}
	case reversed:
		return Outcome{t.ExactNotWinner, RuleReversed}
	case correctWinner:
		return Outcome{t.WinnerNotExact, RuleAdv}
	default:
		return Outcome{t.Nothing, RuleWrong}
	}
}
