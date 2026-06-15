package db

import (
	"context"
	"fmt"
	"time"

	"polla/internal/model"
)

// UpsertMatches inserts or updates fixtures/results in a single transaction.
func (d *DB) UpsertMatches(ctx context.Context, matches []model.Match) error {
	if len(matches) == 0 {
		return nil
	}
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, m := range matches {
		_, err := tx.Exec(ctx, `
			INSERT INTO matches (
				id, utc_date, stage, group_letter, matchday, seq,
				home_team_id, away_team_id, home_team_name, away_team_name,
				status, score_home, score_away, winner, duration, last_synced_at
			) VALUES (
				$1, $2, $3, $4, $5, $6,
				$7, $8, $9, $10,
				$11, $12, $13, $14, $15, now()
			)
			ON CONFLICT (id) DO UPDATE SET
				utc_date = EXCLUDED.utc_date,
				stage = EXCLUDED.stage,
				group_letter = EXCLUDED.group_letter,
				matchday = EXCLUDED.matchday,
				seq = EXCLUDED.seq,
				home_team_id = EXCLUDED.home_team_id,
				away_team_id = EXCLUDED.away_team_id,
				home_team_name = EXCLUDED.home_team_name,
				away_team_name = EXCLUDED.away_team_name,
				-- Preserve admin-entered results: the sync must not overwrite them.
				status = CASE WHEN matches.result_manual THEN matches.status ELSE EXCLUDED.status END,
				score_home = CASE WHEN matches.result_manual THEN matches.score_home ELSE EXCLUDED.score_home END,
				score_away = CASE WHEN matches.result_manual THEN matches.score_away ELSE EXCLUDED.score_away END,
				winner = CASE WHEN matches.result_manual THEN matches.winner ELSE EXCLUDED.winner END,
				duration = CASE WHEN matches.result_manual THEN matches.duration ELSE EXCLUDED.duration END,
				last_synced_at = now()`,
			m.ID, m.UTCDate, m.Stage, nullStr(m.GroupLetter), m.Matchday, m.Seq,
			m.HomeTeamID, m.AwayTeamID, m.HomeTeamName, m.AwayTeamName,
			m.Status, m.ScoreHome, m.ScoreAway, nullStr(m.Winner), nullStr(m.Duration))
		if err != nil {
			return fmt.Errorf("upsert match %d: %w", m.ID, err)
		}
	}
	return tx.Commit(ctx)
}

// ListMatches returns all matches in chronological order (by utc_date, id).
func (d *DB) ListMatches(ctx context.Context) ([]model.Match, error) {
	rows, err := d.Pool.Query(ctx, `
		SELECT id, utc_date, stage, COALESCE(group_letter, ''), COALESCE(matchday, 0), seq,
		       home_team_id, away_team_id, home_team_name, away_team_name,
		       status, score_home, score_away, COALESCE(winner, ''), COALESCE(duration, '')
		FROM matches
		ORDER BY utc_date, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Match
	for rows.Next() {
		var m model.Match
		if err := rows.Scan(
			&m.ID, &m.UTCDate, &m.Stage, &m.GroupLetter, &m.Matchday, &m.Seq,
			&m.HomeTeamID, &m.AwayTeamID, &m.HomeTeamName, &m.AwayTeamName,
			&m.Status, &m.ScoreHome, &m.ScoreAway, &m.Winner, &m.Duration,
		); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// HasLiveMatches reports whether any match kicked off within the last few hours
// but is not yet recorded as FINISHED in our cache. While true, the sync policy
// keeps polling so final results are captured promptly instead of waiting for
// the next match's kickoff. Manually-set results are ignored (no need to poll).
func (d *DB) HasLiveMatches(ctx context.Context, now time.Time) (bool, error) {
	var exists bool
	err := d.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM matches
			WHERE status <> 'FINISHED'
			  AND result_manual = false
			  AND utc_date <= $1
			  AND utc_date >= $1 - INTERVAL '4 hours'
		)`, now).Scan(&exists)
	return exists, err
}

// GetMatch returns a single match by id.
func (d *DB) GetMatch(ctx context.Context, id int64) (model.Match, error) {
	var m model.Match
	err := d.Pool.QueryRow(ctx, `
		SELECT id, utc_date, stage, COALESCE(group_letter, ''), COALESCE(matchday, 0), seq,
		       home_team_id, away_team_id, home_team_name, away_team_name,
		       status, score_home, score_away, COALESCE(winner, ''), COALESCE(duration, '')
		FROM matches WHERE id = $1`, id).Scan(
		&m.ID, &m.UTCDate, &m.Stage, &m.GroupLetter, &m.Matchday, &m.Seq,
		&m.HomeTeamID, &m.AwayTeamID, &m.HomeTeamName, &m.AwayTeamName,
		&m.Status, &m.ScoreHome, &m.ScoreAway, &m.Winner, &m.Duration)
	return m, err
}
