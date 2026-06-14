package db

import (
	"context"
	"time"
)

// SyncState is the single-row bookkeeping for the football-data sync policy.
type SyncState struct {
	LastFullSyncAt  *time.Time
	NextMatchUTC    *time.Time
	NextMatchID     *int64
	NextMatchSynced bool
}

// GetSyncState reads the singleton sync-state row.
func (d *DB) GetSyncState(ctx context.Context) (SyncState, error) {
	var s SyncState
	err := d.Pool.QueryRow(ctx, `
		SELECT last_full_sync_at, next_match_utc, next_match_id, next_match_synced
		FROM api_sync_state WHERE id = 1`).
		Scan(&s.LastFullSyncAt, &s.NextMatchUTC, &s.NextMatchID, &s.NextMatchSynced)
	return s, err
}

// SaveSyncState updates the singleton sync-state row.
func (d *DB) SaveSyncState(ctx context.Context, s SyncState) error {
	_, err := d.Pool.Exec(ctx, `
		UPDATE api_sync_state SET
			last_full_sync_at = $1,
			next_match_utc = $2,
			next_match_id = $3,
			next_match_synced = $4
		WHERE id = 1`,
		s.LastFullSyncAt, s.NextMatchUTC, s.NextMatchID, s.NextMatchSynced)
	return err
}
