package standings

import (
	"reflect"
	"testing"
)

func m(h, a, hg, ag int) Match {
	return Match{HomeTeamID: h, AwayTeamID: a, HomeGoals: hg, AwayGoals: ag, Played: true}
}

func TestComputeCleanTable(t *testing.T) {
	// Team 1 wins all, 2 beats 3&4, 3 beats 4, 4 loses all.
	matches := []Match{
		m(1, 2, 1, 0), m(1, 3, 1, 0), m(1, 4, 1, 0),
		m(2, 3, 1, 0), m(2, 4, 1, 0),
		m(3, 4, 1, 0),
	}
	got := Order(Compute([]int{1, 2, 3, 4}, matches))
	want := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

func TestComputeGoalDiffAndHeadToHead(t *testing.T) {
	// Teams 1, 2 and 4 all finish on 6 points.
	// Team 4 has the best goal difference (+2) → 1st.
	// Teams 1 and 2 tie on points, GD (+1) and GF (4); team 1 beat team 2
	// head-to-head → team 1 ranks above team 2. Team 3 loses everything.
	matches := []Match{
		m(1, 2, 2, 1), // 1 beats 2 (head-to-head)
		m(1, 3, 2, 0),
		m(4, 1, 2, 0),
		m(2, 3, 2, 1),
		m(2, 4, 1, 0),
		m(3, 4, 0, 1),
	}
	got := Order(Compute([]int{1, 2, 3, 4}, matches))
	want := []int{4, 1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

func TestComputeIgnoresUnplayed(t *testing.T) {
	matches := []Match{
		m(1, 2, 3, 0),
		{HomeTeamID: 3, AwayTeamID: 4, HomeGoals: 9, AwayGoals: 0, Played: false},
	}
	rows := Compute([]int{1, 2, 3, 4}, matches)
	byID := map[int]Row{}
	for _, r := range rows {
		byID[r.TeamID] = r
	}
	if byID[1].Points != 3 || byID[1].Played != 1 {
		t.Errorf("team 1 = %+v, want 3 pts / 1 played", byID[1])
	}
	if byID[3].Played != 0 || byID[3].Points != 0 {
		t.Errorf("team 3 should be untouched by an unplayed match, got %+v", byID[3])
	}
}
