// Package football talks to the football-data.org v4 API and keeps a local
// cache of the World Cup fixtures and results in the database, following a
// low-frequency sync policy that respects the free-tier rate limits.
package football

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"polla/internal/model"
)

const defaultBaseURL = "https://api.football-data.org/v4"

// Client is a thin HTTP client for football-data.org.
type Client struct {
	token       string
	competition string
	baseURL     string
	http        *http.Client
}

// NewClient builds a client for the given competition code (e.g. "WC").
func NewClient(token, competition string) *Client {
	return &Client{
		token:       token,
		competition: competition,
		baseURL:     defaultBaseURL,
		http:        &http.Client{Timeout: 20 * time.Second},
	}
}

// apiMatchesResponse mirrors the parts of the /matches payload we use.
type apiMatchesResponse struct {
	Matches []apiMatch `json:"matches"`
}

type apiMatch struct {
	ID       int64     `json:"id"`
	UTCDate  time.Time `json:"utcDate"`
	Status   string    `json:"status"`
	Stage    string    `json:"stage"`
	Group    *string   `json:"group"`
	Matchday *int      `json:"matchday"`
	HomeTeam apiTeam   `json:"homeTeam"`
	AwayTeam apiTeam   `json:"awayTeam"`
	Score    apiScore  `json:"score"`
}

type apiTeam struct {
	ID        *int64  `json:"id"`
	Name      *string `json:"name"`
	ShortName *string `json:"shortName"`
	TLA       *string `json:"tla"`
	Crest     *string `json:"crest"`
}

type scorePair struct {
	Home *int `json:"home"`
	Away *int `json:"away"`
}

type apiScore struct {
	Winner      *string   `json:"winner"`
	Duration    *string   `json:"duration"`
	FullTime    scorePair `json:"fullTime"`
	RegularTime scorePair `json:"regularTime"`
	ExtraTime   scorePair `json:"extraTime"`
	Penalties   scorePair `json:"penalties"`
}

// isShootout reports whether the match was decided by a penalty shootout. In
// that case fullTime folds the shootout into the scoreline (e.g. 1–1 in 90'
// becomes 6–5), which is not the marcador our reglamento scores.
func (s apiScore) isShootout() bool {
	return s.Duration != nil && *s.Duration == "PENALTY_SHOOTOUT"
}

// marcador returns the official result our rules score: the score after regular
// (+ extra) time. For shootouts fullTime includes the shootout, so we rebuild
// from regularTime + extraTime; otherwise fullTime already excludes penalties.
func (s apiScore) marcador() (*int, *int) {
	if s.isShootout() && s.RegularTime.Home != nil && s.RegularTime.Away != nil {
		h := *s.RegularTime.Home + valOr0(s.ExtraTime.Home)
		a := *s.RegularTime.Away + valOr0(s.ExtraTime.Away)
		return &h, &a
	}
	return s.FullTime.Home, s.FullTime.Away
}

// resolveWinner returns the advancing side in football-data's HOME_TEAM/AWAY_TEAM
// vocabulary. After a shootout the marcador is a draw, so the scoreline can no
// longer imply the winner — we read it from the shootout, preferring the
// explicit winner field, then the penalties tally, then the fullTime aggregate
// (which always favours the team that advanced).
func (s apiScore) resolveWinner() string {
	if s.Winner != nil && *s.Winner != "" {
		return *s.Winner
	}
	if s.isShootout() {
		if w := sideOf(s.Penalties); w != "" {
			return w
		}
		return sideOf(s.FullTime)
	}
	return ""
}

// sideOf returns HOME_TEAM/AWAY_TEAM for a decisive pair, or "" for a tie/unknown.
func sideOf(p scorePair) string {
	if p.Home == nil || p.Away == nil || *p.Home == *p.Away {
		return ""
	}
	if *p.Home > *p.Away {
		return "HOME_TEAM"
	}
	return "AWAY_TEAM"
}

func valOr0(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// Matches fetches all matches for the competition and returns them as domain
// models along with the distinct set of (non-placeholder) teams seen.
func (c *Client) Matches(ctx context.Context) ([]model.Match, []model.Team, error) {
	url := fmt.Sprintf("%s/competitions/%s/matches", c.baseURL, c.competition)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("build request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("X-Auth-Token", c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("call football-data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, nil, fmt.Errorf("football-data returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload apiMatchesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, nil, fmt.Errorf("decode response: %w", err)
	}

	matches := make([]model.Match, 0, len(payload.Matches))
	teamsByID := map[int64]model.Team{}
	for _, am := range payload.Matches {
		matches = append(matches, am.toModel())
		collectTeam(teamsByID, am.HomeTeam, am.Group)
		collectTeam(teamsByID, am.AwayTeam, am.Group)
	}

	teams := make([]model.Team, 0, len(teamsByID))
	for _, t := range teamsByID {
		teams = append(teams, t)
	}
	return matches, teams, nil
}

func (am apiMatch) toModel() model.Match {
	m := model.Match{
		ID:           am.ID,
		UTCDate:      am.UTCDate,
		Stage:        am.Stage,
		GroupLetter:  groupLetter(am.Group),
		Status:       am.Status,
		HomeTeamID:   am.HomeTeam.ID,
		AwayTeamID:   am.AwayTeam.ID,
		HomeTeamName: teamLabel(am.HomeTeam),
		AwayTeamName: teamLabel(am.AwayTeam),
		Winner:       am.Score.resolveWinner(),
		Duration:     deref(am.Score.Duration),
	}
	m.ScoreHome, m.ScoreAway = am.Score.marcador()
	if am.Matchday != nil {
		m.Matchday = *am.Matchday
	}
	return m
}

func collectTeam(into map[int64]model.Team, t apiTeam, group *string) {
	if t.ID == nil {
		return // placeholder team (knockout slot not yet decided)
	}
	into[*t.ID] = model.Team{
		ID:          *t.ID,
		Name:        deref(t.Name),
		ShortName:   deref(t.ShortName),
		TLA:         deref(t.TLA),
		CrestURL:    deref(t.Crest),
		GroupLetter: groupLetter(group),
	}
}

// teamLabel returns the team name, or a placeholder for an undecided knockout
// slot.
func teamLabel(t apiTeam) string {
	if t.Name != nil && *t.Name != "" {
		return *t.Name
	}
	return "Por definir"
}

// groupLetter converts "GROUP_A" into "A"; returns "" when there is no group.
func groupLetter(g *string) string {
	if g == nil {
		return ""
	}
	return strings.TrimPrefix(*g, "GROUP_")
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
