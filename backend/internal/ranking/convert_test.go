package ranking

import (
	"testing"

	"polla/internal/model"
)

func ptr(n int) *int { return &n }

// knockoutMatch builds a finished LAST_32 match with the given real score and
// the raw winner string as stored in the DB.
func knockoutMatch(home, away int, winner string) model.Match {
	return model.Match{
		Stage:     "LAST_32",
		Status:    "FINISHED",
		ScoreHome: ptr(home),
		ScoreAway: ptr(away),
		Winner:    winner,
	}
}

func pred(home, away int) model.CardPrediction {
	return model.CardPrediction{Home: ptr(home), Away: ptr(away)}
}

// TestKnockoutWinnerVocabulary guards the football-data ("*_TEAM") vs admin
// ("*_WIN") winner mismatch that scored knockout matches as if the winner were
// unknown. Both screenshots: real 0-1, winner synced as "AWAY_TEAM".
func TestKnockoutWinnerVocabulary(t *testing.T) {
	cases := []struct {
		name   string
		m      model.Match
		p      model.CardPrediction
		want   int
	}{
		{"exact + winner, synced AWAY_TEAM", knockoutMatch(0, 1, "AWAY_TEAM"), pred(0, 1), 7},
		{"exact + winner, manual AWAY_WIN", knockoutMatch(0, 1, "AWAY_WIN"), pred(0, 1), 7},
		{"winner only, synced AWAY_TEAM", knockoutMatch(0, 1, "AWAY_TEAM"), pred(1, 2), 3},
		{"winner only, empty winner string", knockoutMatch(0, 1, ""), pred(1, 2), 3},
		{"home winner, synced HOME_TEAM", knockoutMatch(2, 0, "HOME_TEAM"), pred(3, 1), 3},
		{"nothing right", knockoutMatch(0, 1, "AWAY_TEAM"), pred(2, 0), 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := MatchPoints(c.m, c.p); got != c.want {
				t.Errorf("MatchPoints = %d, want %d", got, c.want)
			}
		})
	}
}
