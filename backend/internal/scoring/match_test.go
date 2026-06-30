package scoring

import "testing"

func filled(h, a int) Prediction { return Prediction{Home: h, Away: a, Filled: true} }

func TestScoreGroupMatch(t *testing.T) {
	cases := []struct {
		name string
		pred Prediction
		res  Result
		want int
	}{
		{"empty", Prediction{}, Result{Home: 1, Away: 0, Winner: SideHome}, 0},
		{"half filled", Prediction{Home: 2, Filled: false}, Result{Home: 2, Away: 1, Winner: SideHome}, 0},
		{"exact win", filled(2, 1), Result{Home: 2, Away: 1, Winner: SideHome}, 7},
		{"exact draw", filled(1, 1), Result{Home: 1, Away: 1, Winner: SideNone}, 7},
		{"reversed score", filled(2, 1), Result{Home: 1, Away: 2, Winner: SideAway}, 5},
		{"reversed score other way", filled(0, 3), Result{Home: 3, Away: 0, Winner: SideHome}, 5},
		{"right winner wrong score", filled(2, 1), Result{Home: 3, Away: 1, Winner: SideHome}, 4},
		{"right winner away", filled(0, 1), Result{Home: 1, Away: 3, Winner: SideAway}, 4},
		{"predicted draw actual draw wrong score", filled(1, 1), Result{Home: 2, Away: 2, Winner: SideNone}, 3},
		{"predicted win but draw", filled(2, 1), Result{Home: 0, Away: 0, Winner: SideNone}, 1},
		{"predicted draw but win", filled(1, 1), Result{Home: 2, Away: 1, Winner: SideHome}, 1},
		{"wrong winner not reversed", filled(2, 0), Result{Home: 0, Away: 1, Winner: SideAway}, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ScoreGroupMatch(c.pred, c.res); got != c.want {
				t.Errorf("ScoreGroupMatch(%+v, %+v) = %d, want %d", c.pred, c.res, got, c.want)
			}
		})
	}
}

func TestScoreKnockoutMatch(t *testing.T) {
	draw := func(h, a int, pen Side) Prediction {
		return Prediction{Home: h, Away: a, Filled: true, PenaltyWinner: pen}
	}
	cases := []struct {
		name  string
		pred  Prediction
		res   Result
		tiers KnockoutPoints
		want  int
	}{
		{"empty", Prediction{}, Result{Home: 1, Away: 0, Winner: SideHome}, KnockoutTiers, 0},
		{"exact and winner", filled(2, 1), Result{Home: 2, Away: 1, Winner: SideHome}, KnockoutTiers, 7},
		{"exact draw correct penalty", draw(1, 1, SideHome), Result{Home: 1, Away: 1, Winner: SideHome}, KnockoutTiers, 7},
		{"exact draw wrong penalty", draw(1, 1, SideAway), Result{Home: 1, Away: 1, Winner: SideHome}, KnockoutTiers, 5},
		{"exact draw no penalty pick", filled(1, 1), Result{Home: 1, Away: 1, Winner: SideHome}, KnockoutTiers, 5},
		{"reversed score", filled(1, 2), Result{Home: 2, Away: 1, Winner: SideHome}, KnockoutTiers, 5},
		{"winner right wrong score", filled(3, 0), Result{Home: 2, Away: 1, Winner: SideHome}, KnockoutTiers, 3},
		{"nothing right", filled(2, 1), Result{Home: 0, Away: 3, Winner: SideAway}, KnockoutTiers, 1},
		{"final exact and winner", filled(1, 0), Result{Home: 1, Away: 0, Winner: SideHome}, FinalTiers, 10},
		{"final exact wrong penalty", draw(2, 2, SideHome), Result{Home: 2, Away: 2, Winner: SideAway}, FinalTiers, 8},
		{"final winner wrong score", filled(2, 0), Result{Home: 1, Away: 0, Winner: SideHome}, FinalTiers, 6},
		{"final nothing", filled(0, 1), Result{Home: 3, Away: 0, Winner: SideHome}, FinalTiers, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ScoreKnockoutMatch(c.pred, c.res, c.tiers); got != c.want {
				t.Errorf("ScoreKnockoutMatch(%+v, %+v) = %d, want %d", c.pred, c.res, got, c.want)
			}
		})
	}
}

func TestScoreBracketMatchup(t *testing.T) {
	hit := BracketPrediction{HomeTeamID: 10, AwayTeamID: 20, Filled: true}
	actual := BracketActual{HomeTeamID: 10, AwayTeamID: 20}
	if got := ScoreBracketMatchup(hit, actual, BracketMatchupPoints); got != 5 {
		t.Errorf("correct matchup = %d, want 5", got)
	}
	if got := ScoreBracketMatchup(hit, actual, BracketFinalMatchupPoints); got != 7 {
		t.Errorf("correct final matchup = %d, want 7", got)
	}
	reversed := BracketPrediction{HomeTeamID: 20, AwayTeamID: 10, Filled: true}
	if got := ScoreBracketMatchup(reversed, actual, BracketMatchupPoints); got != 0 {
		t.Errorf("reversed order = %d, want 0 (order matters)", got)
	}
	if got := ScoreBracketMatchup(BracketPrediction{}, actual, BracketMatchupPoints); got != 0 {
		t.Errorf("empty = %d, want 0", got)
	}
}
