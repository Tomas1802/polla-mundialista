package db

import (
	"context"
	"fmt"

	"polla/internal/model"
)

// UpsertTeams inserts or updates the given teams in a single transaction.
func (d *DB) UpsertTeams(ctx context.Context, teams []model.Team) error {
	if len(teams) == 0 {
		return nil
	}
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, t := range teams {
		_, err := tx.Exec(ctx, `
			INSERT INTO teams (id, name, short_name, tla, crest_url, group_letter)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				short_name = EXCLUDED.short_name,
				tla = EXCLUDED.tla,
				crest_url = EXCLUDED.crest_url,
				group_letter = COALESCE(EXCLUDED.group_letter, teams.group_letter)`,
			t.ID, t.Name, t.ShortName, t.TLA, t.CrestURL, nullStr(t.GroupLetter))
		if err != nil {
			return fmt.Errorf("upsert team %d: %w", t.ID, err)
		}
	}
	return tx.Commit(ctx)
}

// ListTeams returns all teams keyed by id.
func (d *DB) ListTeams(ctx context.Context) (map[int64]model.Team, error) {
	rows, err := d.Pool.Query(ctx, `
		SELECT id, name, short_name, tla, crest_url, COALESCE(group_letter, '')
		FROM teams`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[int64]model.Team{}
	for rows.Next() {
		var t model.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.ShortName, &t.TLA, &t.CrestURL, &t.GroupLetter); err != nil {
			return nil, err
		}
		out[t.ID] = t
	}
	return out, rows.Err()
}
