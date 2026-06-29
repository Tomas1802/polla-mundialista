package football

import "testing"

func p(n int) *int { return &n }
func s(v string) *string { return &v }

// TestMarcadorAndWinner guards the penalty-shootout parsing: fullTime folds in
// the shootout (e.g. Germany 1–1 Paraguay, won on penalties, reported 5–6), but
// our rules score the 90'/120' marcador and need the advancing side.
func TestMarcadorAndWinner(t *testing.T) {
	cases := []struct {
		name             string
		score            apiScore
		wantH, wantA     int
		wantWinner       string
	}{
		{
			name: "shootout uses regular time, winner from fullTime when winner+penalties tie",
			score: apiScore{
				Duration:    s("PENALTY_SHOOTOUT"),
				FullTime:    scorePair{p(5), p(6)},
				RegularTime: scorePair{p(1), p(1)},
				ExtraTime:   scorePair{p(0), p(0)},
				Penalties:   scorePair{p(5), p(5)},
			},
			wantH: 1, wantA: 1, wantWinner: "AWAY_TEAM",
		},
		{
			name: "shootout winner from penalties tally",
			score: apiScore{
				Duration:    s("PENALTY_SHOOTOUT"),
				FullTime:    scorePair{p(4), p(5)},
				RegularTime: scorePair{p(2), p(2)},
				Penalties:   scorePair{p(3), p(4)},
			},
			wantH: 2, wantA: 2, wantWinner: "AWAY_TEAM",
		},
		{
			name: "regular match keeps fullTime and explicit winner",
			score: apiScore{
				Duration: s("REGULAR"),
				Winner:   s("HOME_TEAM"),
				FullTime: scorePair{p(2), p(0)},
			},
			wantH: 2, wantA: 0, wantWinner: "HOME_TEAM",
		},
		{
			name: "extra-time win counts extra-time goals, not penalties",
			score: apiScore{
				Duration:    s("EXTRA_TIME"),
				Winner:      s("HOME_TEAM"),
				FullTime:    scorePair{p(2), p(1)},
				RegularTime: scorePair{p(1), p(1)},
				ExtraTime:   scorePair{p(1), p(0)},
			},
			wantH: 2, wantA: 1, wantWinner: "HOME_TEAM",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h, a := c.score.marcador()
			if h == nil || a == nil || *h != c.wantH || *a != c.wantA {
				t.Errorf("marcador = %v-%v, want %d-%d", h, a, c.wantH, c.wantA)
			}
			if w := c.score.resolveWinner(); w != c.wantWinner {
				t.Errorf("resolveWinner = %q, want %q", w, c.wantWinner)
			}
		})
	}
}
