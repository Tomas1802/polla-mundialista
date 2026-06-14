package scoring

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
	if !p.Filled {
		return 0
	}
	if p.Home == r.Home && p.Away == r.Away {
		return 7
	}
	// "Acierta el marcador, pero con los equipos al revés" — e.g. predicted 2-1
	// when the result was 1-2. Only meaningful when there is a winner; a draw
	// reversed is identical to the exact score and is already handled above.
	if p.Home == r.Away && p.Away == r.Home && !r.isDraw() {
		return 5
	}
	predDraw := p.Home == p.Away
	if !r.isDraw() && !predDraw && sign(p.Home-p.Away) == sign(r.Home-r.Away) {
		return 4
	}
	if r.isDraw() && predDraw {
		return 3
	}
	return 1
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
	if !p.Filled {
		return 0
	}
	exact := p.Home == r.Home && p.Away == r.Away
	correctWinner := r.Winner != SideNone && p.winner() == r.Winner
	switch {
	case exact && correctWinner:
		return t.ExactAndWinner
	case exact && !correctWinner:
		return t.ExactNotWinner
	case !exact && correctWinner:
		return t.WinnerNotExact
	default:
		return t.Nothing
	}
}
