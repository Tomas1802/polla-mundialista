package ranking

import (
	"context"
	"sort"
)

// TeamMember is a player's contribution to their team (their best card).
type TeamMember struct {
	PlayerID int64  `json:"playerId"`
	Name     string `json:"name"`
	Points   int    `json:"points"`
}

// TeamEntry is a team's place in the team ranking.
type TeamEntry struct {
	Name    string       `json:"name"`
	Points  int          `json:"points"`
	Rank    int          `json:"rank"`
	Members []TeamMember `json:"members"`
}

// TeamStandings ranks teams by the sum of each member's BEST card, and also
// returns the players that belong to no team ("sin equipo").
func (s *Service) TeamStandings(ctx context.Context) ([]TeamEntry, []TeamMember, error) {
	cardEntries, err := s.Standings(ctx)
	if err != nil {
		return nil, nil, err
	}
	// Best card points per player.
	best := map[int64]int{}
	for _, e := range cardEntries {
		if cur, ok := best[e.PlayerID]; !ok || e.Points > cur {
			best[e.PlayerID] = e.Points
		}
	}

	playerTeams, err := s.store.ListPlayerTeams(ctx)
	if err != nil {
		return nil, nil, err
	}

	type acc struct {
		name    string
		members []TeamMember
		total   int
	}
	teams := map[int64]*acc{}
	var sinEquipo []TeamMember

	for _, pt := range playerTeams {
		m := TeamMember{PlayerID: pt.PlayerID, Name: pt.PlayerName, Points: best[pt.PlayerID]}
		if pt.TeamID == nil {
			sinEquipo = append(sinEquipo, m)
			continue
		}
		a := teams[*pt.TeamID]
		if a == nil {
			a = &acc{name: pt.TeamName}
			teams[*pt.TeamID] = a
		}
		a.members = append(a.members, m)
		a.total += m.Points
	}

	out := make([]TeamEntry, 0, len(teams))
	for _, a := range teams {
		sort.SliceStable(a.members, func(i, j int) bool { return a.members[i].Points > a.members[j].Points })
		out = append(out, TeamEntry{Name: a.name, Points: a.total, Members: a.members})
	}
	// Rank: most points first; ties broken by team name (sequential places).
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Points != out[j].Points {
			return out[i].Points > out[j].Points
		}
		return out[i].Name < out[j].Name
	})
	for i := range out {
		out[i].Rank = i + 1
	}

	sort.SliceStable(sinEquipo, func(i, j int) bool { return sinEquipo[i].Name < sinEquipo[j].Name })
	return out, sinEquipo, nil
}
