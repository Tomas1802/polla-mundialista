package db

import "context"

// GetEditLockMatchID returns the cutoff match id: matches up to and including it
// (chronologically) are not editable. nil means no lock.
func (d *DB) GetEditLockMatchID(ctx context.Context) (*int64, error) {
	var id *int64
	err := d.Pool.QueryRow(ctx, `SELECT edit_lock_until_match_id FROM app_settings WHERE id = 1`).Scan(&id)
	return id, err
}

// SetEditLockMatchID sets (or clears, with nil) the edit-lock cutoff match.
func (d *DB) SetEditLockMatchID(ctx context.Context, matchID *int64) error {
	_, err := d.Pool.Exec(ctx, `UPDATE app_settings SET edit_lock_until_match_id = $1 WHERE id = 1`, matchID)
	return err
}
