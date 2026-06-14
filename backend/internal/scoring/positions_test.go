package scoring

import "testing"

func TestScoreGroupPositions(t *testing.T) {
	real := GroupOrder{10, 20, 30, 40}
	cases := []struct {
		name string
		pred GroupOrder
		want int
	}{
		{"exact", GroupOrder{10, 20, 30, 40}, 7},
		{"top two right, 3-4 swapped", GroupOrder{10, 20, 40, 30}, 4},
		{"only first right", GroupOrder{10, 30, 40, 20}, 3},
		{"only third right", GroupOrder{20, 40, 30, 10}, 3},
		{"nothing right", GroupOrder{40, 30, 20, 10}, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ScoreGroupPositions(c.pred, real); got != c.want {
				t.Errorf("ScoreGroupPositions(%v) = %d, want %d", c.pred, got, c.want)
			}
		})
	}
}
