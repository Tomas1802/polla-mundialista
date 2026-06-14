package scoring

// GroupOrder is the final ranking of a group's four teams, position 1 to 4,
// given as team ids. For predictions it is DERIVED from the user's predicted
// group-match scores (see package standings); for the real table it comes from
// the official results.
type GroupOrder [4]int

// ScoreGroupPositions implements section 2 of the reglamento by comparing a
// predicted group order against the real one.
//
// INTERPRETATION (pending organizer confirmation — kept isolated and easy to
// tune): because each group has four fixed teams, several reglamento rows are
// unreachable or trivial (e.g. "acierta 1,2,3 en orden" forces the 4th, and
// "acierta todos los clasificados en cualquier orden" is always true). The
// finer 2-point rows (only-1st, only-2nd, only-3rd-that-qualifies) and the
// cross-group "best third" qualification are intentionally simplified here:
//
//	7  exact order 1-2-3-4
//	4  positions 1 and 2 exactly right (3rd/4th differ)
//	3  at least one position exactly right (but not enough for 4 or 7)
//	1  no position right
//
// A group with no recorded prediction must be guarded by the caller; this
// function always assumes a real attempt was derived.
func ScoreGroupPositions(pred, real GroupOrder) int {
	if pred == real {
		return 7
	}
	if pred[0] == real[0] && pred[1] == real[1] {
		return 4
	}
	for i := range pred {
		if pred[i] == real[i] {
			return 3
		}
	}
	return 1
}
