package scoring

// GroupOrder is the final ranking of a group's four teams, position 1 to 4,
// given as team ids. For predictions it is DERIVED from the user's predicted
// group-match scores (see package standings); for the real table it comes from
// the official results.
type GroupOrder [4]int

// ScoreGroupPositions implements section 2 of the reglamento by comparing a
// predicted group order against the real one. Positions 1-3 are the qualifiers;
// the 4th is non-qualifying. The rules are a priority ladder — the first that
// holds wins:
//
//	7  todas las posiciones en orden correcto
//	4  1er y 2do puesto en orden correcto
//	3  1er, 2do y 3er puesto acertados (en cualquier orden)
//	2  acierta al menos un clasificado en su posición correspondiente
//	1  no cumple ninguna de las anteriores
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
	if sameTopThree(pred, real) {
		return 3
	}
	for i := 0; i < 3; i++ {
		if pred[i] == real[i] {
			return 2
		}
	}
	return 1
}

// sameTopThree reports whether the predicted top three (qualifiers) are the
// same set of teams as the real top three, regardless of internal order.
func sameTopThree(pred, real GroupOrder) bool {
	seen := map[int]bool{real[0]: true, real[1]: true, real[2]: true}
	for i := 0; i < 3; i++ {
		if !seen[pred[i]] {
			return false
		}
	}
	return true
}
