package scoring

// Side identifies a team within a match by its position in the fixture.
type Side string

const (
	SideHome Side = "HOME"
	SideAway Side = "AWAY"
	// SideNone means "no winner": a group-stage draw, or a knockout prediction
	// the user left as a draw without choosing who advances on penalties.
	SideNone Side = ""
)

// Prediction is a user's forecast for a single match.
type Prediction struct {
	Home int
	Away int
	// Filled reports whether the user entered BOTH scores. Per the reglamento,
	// an empty or half-filled prediction always scores 0.
	Filled bool
	// PenaltyWinner only matters for knockout matches the user predicted as a
	// draw: it records which team they think advances on penalties. It is
	// SideNone for group matches and for non-draw knockout predictions.
	PenaltyWinner Side
}

// Result is the official outcome of a match.
type Result struct {
	Home int
	Away int
	// Winner is the team that advances. For group matches it may be SideNone
	// (a draw). For knockout matches it is always SideHome or SideAway because
	// penalties break ties. This maps directly to football-data.org's
	// score.winner (DRAW / HOME_WIN / AWAY_WIN) combined with the penalty
	// shootout outcome.
	Winner Side
}

// winner returns the team the prediction implies will advance.
func (p Prediction) winner() Side {
	switch {
	case p.Home > p.Away:
		return SideHome
	case p.Away > p.Home:
		return SideAway
	default:
		return p.PenaltyWinner // a predicted draw → the penalty pick (maybe none)
	}
}

func (r Result) isDraw() bool { return r.Home == r.Away }

func sign(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	default:
		return 0
	}
}
