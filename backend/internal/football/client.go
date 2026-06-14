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

type apiScore struct {
	Winner   *string `json:"winner"`
	Duration *string `json:"duration"`
	FullTime struct {
		Home *int `json:"home"`
		Away *int `json:"away"`
	} `json:"fullTime"`
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
		ScoreHome:    am.Score.FullTime.Home,
		ScoreAway:    am.Score.FullTime.Away,
		Winner:       deref(am.Score.Winner),
		Duration:     deref(am.Score.Duration),
	}
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
