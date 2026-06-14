// Package standings computes a group table from a set of match scores, ordered
// with the FIFA World Cup group-ranking criteria. It is used twice: on a user's
// predicted scores to derive their predicted table, and on the official scores
// to build the real table. Comparing the two yields section-2 points.
//
// FIFA group ranking criteria (in order):
//  1. Points in all group matches
//  2. Goal difference in all group matches
//  3. Goals scored in all group matches
//  4. Points in matches between the tied teams (head-to-head)
//  5. Goal difference in those matches
//  6. Goals scored in those matches
//  7. Fair play, then 8. drawing of lots
//
// Criteria 7-8 cannot be derived from scores alone, so ties that survive
// head-to-head are broken deterministically by ascending team id. This only
// affects display order for genuinely-tied teams and is documented for the
// organizer.
package standings

import "sort"

// Match is a single group game with a final score.
type Match struct {
	HomeTeamID int
	AwayTeamID int
	HomeGoals  int
	AwayGoals  int
	Played     bool // unplayed matches do not contribute to the table
}

// Row is one team's aggregated record within a group.
type Row struct {
	TeamID       int
	Played       int
	Won          int
	Drawn        int
	Lost         int
	GoalsFor     int
	GoalsAgainst int
	Points       int
}

// GoalDiff is goals for minus goals against.
func (r Row) GoalDiff() int { return r.GoalsFor - r.GoalsAgainst }

// Compute returns the group table for teamIDs, ordered best-first.
func Compute(teamIDs []int, matches []Match) []Row {
	out := aggregate(teamIDs, matches)

	// Primary order: overall points, GD, GF; final fallback ascending team id.
	sortByOverall(out)

	// Break ties among teams equal on (points, GD, GF) using head-to-head.
	for i := 0; i < len(out); {
		j := i + 1
		for j < len(out) && sameOverall(out[i], out[j]) {
			j++
		}
		if j-i > 1 {
			resolveHeadToHead(out[i:j], matches)
		}
		i = j
	}
	return out
}

// Order returns just the ordered team ids.
func Order(rows []Row) []int {
	ids := make([]int, len(rows))
	for i, r := range rows {
		ids[i] = r.TeamID
	}
	return ids
}

// aggregate builds an unsorted table from the matches involving teamIDs.
func aggregate(teamIDs []int, matches []Match) []Row {
	rows := make(map[int]*Row, len(teamIDs))
	for _, id := range teamIDs {
		rows[id] = &Row{TeamID: id}
	}
	for _, m := range matches {
		h, a := rows[m.HomeTeamID], rows[m.AwayTeamID]
		if !m.Played || h == nil || a == nil {
			continue
		}
		h.Played++
		a.Played++
		h.GoalsFor += m.HomeGoals
		h.GoalsAgainst += m.AwayGoals
		a.GoalsFor += m.AwayGoals
		a.GoalsAgainst += m.HomeGoals
		switch {
		case m.HomeGoals > m.AwayGoals:
			h.Won++
			h.Points += 3
			a.Lost++
		case m.HomeGoals < m.AwayGoals:
			a.Won++
			a.Points += 3
			h.Lost++
		default:
			h.Drawn++
			a.Drawn++
			h.Points++
			a.Points++
		}
	}
	out := make([]Row, 0, len(teamIDs))
	for _, id := range teamIDs {
		out = append(out, *rows[id])
	}
	return out
}

func sortByOverall(rows []Row) {
	sort.SliceStable(rows, func(i, j int) bool {
		if c := overallCompare(rows[i], rows[j]); c != 0 {
			return c > 0
		}
		return rows[i].TeamID < rows[j].TeamID
	})
}

func overallCompare(a, b Row) int {
	if a.Points != b.Points {
		return a.Points - b.Points
	}
	if a.GoalDiff() != b.GoalDiff() {
		return a.GoalDiff() - b.GoalDiff()
	}
	return a.GoalsFor - b.GoalsFor
}

func sameOverall(a, b Row) bool {
	return a.Points == b.Points && a.GoalDiff() == b.GoalDiff() && a.GoalsFor == b.GoalsFor
}

// resolveHeadToHead reorders a run of teams tied on overall criteria using only
// the matches played among themselves. It does not recurse: any team still tied
// after head-to-head keeps the ascending-team-id fallback already in place.
func resolveHeadToHead(run []Row, matches []Match) {
	inRun := make(map[int]bool, len(run))
	ids := make([]int, len(run))
	for i, r := range run {
		inRun[r.TeamID] = true
		ids[i] = r.TeamID
	}
	mini := make([]Match, 0, len(matches))
	for _, m := range matches {
		if m.Played && inRun[m.HomeTeamID] && inRun[m.AwayTeamID] {
			mini = append(mini, m)
		}
	}
	h2h := aggregate(ids, mini)
	sortByOverall(h2h)
	rank := make(map[int]int, len(h2h))
	for i, r := range h2h {
		rank[r.TeamID] = i
	}
	sort.SliceStable(run, func(i, j int) bool {
		return rank[run[i].TeamID] < rank[run[j].TeamID]
	})
}
