// Package model holds the core domain types shared across the backend layers.
package model

import "time"

// Team is a national team. IDs mirror football-data.org team ids.
type Team struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	ShortName   string `json:"shortName"`
	TLA         string `json:"tla"`
	CrestURL    string `json:"crest"`
	GroupLetter string `json:"group"` // "A".."L" or ""
}

// Match is a fixture (and result, once played) cached from football-data.org.
// Team ids are pointers because knockout matches have no teams until they are
// decided; HomeTeamName/AwayTeamName always carry a human label (a team name or
// a placeholder like "Winner Group A").
type Match struct {
	ID           int64     `json:"id"`
	UTCDate      time.Time `json:"utcDate"`
	Stage        string    `json:"stage"`
	GroupLetter  string    `json:"group"` // "A".."L" or ""
	Matchday     int       `json:"matchday"`
	Seq          int       `json:"seq"` // chronological index, assigned on sync
	HomeTeamID   *int64    `json:"homeTeamId"`
	AwayTeamID   *int64    `json:"awayTeamId"`
	HomeTeamName string    `json:"homeTeamName"`
	AwayTeamName string    `json:"awayTeamName"`
	Status       string    `json:"status"`
	ScoreHome    *int      `json:"scoreHome"`
	ScoreAway    *int      `json:"scoreAway"`
	Winner       string    `json:"winner"`   // HOME_WIN | AWAY_WIN | DRAW | ""
	Duration     string    `json:"duration"` // REGULAR | EXTRA_TIME | PENALTY_SHOOTOUT | ""
}

// Finished reports whether the match has an official result.
func (m Match) Finished() bool { return m.Status == "FINISHED" && m.ScoreHome != nil && m.ScoreAway != nil }

// Started reports whether the match is underway or already over, in which case
// its marcador must no longer be editable regardless of the cached kickoff time.
func (m Match) Started() bool {
	switch m.Status {
	case "IN_PLAY", "PAUSED", "FINISHED", "SUSPENDED", "AWARDED":
		return true
	default:
		return false
	}
}

// User is a participant. Identity is the verified phone number.
type User struct {
	ID           int64  `json:"id"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	FirebaseUID  string `json:"-"`
	DisplayName  string `json:"displayName"`
	Timezone     string `json:"timezone"`
	IsAdmin      bool   `json:"isAdmin"`
	PlayerID     *int64 `json:"playerId"`
	PlayerName   string `json:"playerName"`
	SessionEpoch int    `json:"-"`
}

// Player is a participant in the pool. A player owns one to three cards and
// authenticates with a 4-digit PIN.
type Player struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	CardCount     int    `json:"cardCount"`
	PinHash       string `json:"-"`
	HasPin        bool   `json:"-"`
	MustChangePin bool   `json:"mustChangePin"`
	SessionEpoch  int    `json:"-"`
}

// Card (cartón) is one bet sheet belonging to a player.
type Card struct {
	ID       int64  `json:"id"`
	PlayerID int64  `json:"playerId"`
	CardNo   int    `json:"cardNo"`
	Label    string `json:"label"`
}

// CardPrediction is a card's marcador for one match.
type CardPrediction struct {
	CardID        int64  `json:"-"`
	MatchID       int64  `json:"matchId"`
	Home          *int   `json:"home"`
	Away          *int   `json:"away"`
	PenaltyWinner string `json:"penaltyWinner"`
}

// Filled reports whether both scores were entered.
func (p CardPrediction) Filled() bool { return p.Home != nil && p.Away != nil }
