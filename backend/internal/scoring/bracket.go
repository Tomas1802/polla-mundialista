package scoring

// Bracket points from the "acierta los equipos que se enfrentan" rows.
const (
	BracketMatchupPoints      = 5 // sections 3 (dieciseisavos..semifinales)
	BracketFinalMatchupPoints = 7 // section 4 (final)
)

// BracketPrediction is a forecast of which two teams meet in a given knockout
// slot. Teams are identified by their football-data.org team id.
type BracketPrediction struct {
	HomeTeamID int
	AwayTeamID int
	Filled     bool
}

// BracketActual is the real pairing for that slot once it is known.
type BracketActual struct {
	HomeTeamID int
	AwayTeamID int
}

// ScoreBracketMatchup awards points for correctly predicting the two teams that
// meet in a knockout slot, in the correct order, as in the top rows of
// sections 3 and 4. The reglamento defines no partial credit for this row, so a
// wrong pairing scores 0.
//
// PENDING CONFIRMATION (reglamento ambiguity #1): the document does not state
// whether these matchup points are ADDED to the marcador points of the same
// match. This package returns them separately; whether the caller sums them is
// a single decision made in the ranking layer.
func ScoreBracketMatchup(p BracketPrediction, a BracketActual, points int) int {
	if !p.Filled {
		return 0
	}
	if p.HomeTeamID == a.HomeTeamID && p.AwayTeamID == a.AwayTeamID {
		return points
	}
	return 0
}
