package db

import "context"

// UpsertPoolTeam inserts a team by name (or returns the existing id).
func (d *DB) UpsertPoolTeam(ctx context.Context, name string) (int64, error) {
	var id int64
	err := d.Pool.QueryRow(ctx, `
		INSERT INTO pool_teams (name) VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, name).Scan(&id)
	return id, err
}

// ClearTeamAssignments unlinks every player from any team (used before a
// re-import so stale assignments don't linger).
func (d *DB) ClearTeamAssignments(ctx context.Context) error {
	_, err := d.Pool.Exec(ctx, `UPDATE players SET pool_team_id = NULL`)
	return err
}

// AssignPlayerToTeam links a player (by name) to a team. Returns whether a
// player with that name existed.
func (d *DB) AssignPlayerToTeam(ctx context.Context, playerName string, teamID int64) (bool, error) {
	tag, err := d.Pool.Exec(ctx, `UPDATE players SET pool_team_id = $2 WHERE name = $1`, playerName, teamID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// PlayerTeam pairs a player with their team (TeamID nil = sin equipo).
type PlayerTeam struct {
	PlayerID   int64
	PlayerName string
	TeamID     *int64
	TeamName   string
}

// ListPlayerTeams returns every player with their team (if any).
func (d *DB) ListPlayerTeams(ctx context.Context) ([]PlayerTeam, error) {
	rows, err := d.Pool.Query(ctx, `
		SELECT p.id, p.name, p.pool_team_id, COALESCE(t.name, '')
		FROM players p
		LEFT JOIN pool_teams t ON t.id = p.pool_team_id
		ORDER BY p.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PlayerTeam
	for rows.Next() {
		var pt PlayerTeam
		if err := rows.Scan(&pt.PlayerID, &pt.PlayerName, &pt.TeamID, &pt.TeamName); err != nil {
			return nil, err
		}
		out = append(out, pt)
	}
	return out, rows.Err()
}
