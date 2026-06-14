package db

import (
	"context"

	"polla/internal/model"
)

const userColumns = `id, COALESCE(phone, ''), COALESCE(email, ''), firebase_uid, display_name, timezone, is_admin, player_id, session_epoch`

func scanUser(row interface {
	Scan(dest ...any) error
}) (model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.Phone, &u.Email, &u.FirebaseUID, &u.DisplayName, &u.Timezone, &u.IsAdmin, &u.PlayerID, &u.SessionEpoch)
	return u, err
}

// SetUserPlayer links a user to the player they are in the pool.
func (d *DB) SetUserPlayer(ctx context.Context, userID, playerID int64) error {
	_, err := d.Pool.Exec(ctx, `UPDATE users SET player_id = $2 WHERE id = $1`, userID, playerID)
	return err
}

// EmailExists reports whether a user with the given email is already registered.
func (d *DB) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := d.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email) = LOWER($1))`, email).Scan(&exists)
	return exists, err
}

// SetUserAdmin grants or revokes admin rights.
func (d *DB) SetUserAdmin(ctx context.Context, id int64, isAdmin bool) error {
	_, err := d.Pool.Exec(ctx, `UPDATE users SET is_admin = $2 WHERE id = $1`, id, isAdmin)
	return err
}

// UpsertUserByFirebase creates the user on first login or returns the existing
// one, keyed by the stable Firebase UID. The contact value is the email
// (passwordless email-link login).
func (d *DB) UpsertUserByFirebase(ctx context.Context, firebaseUID, email string) (model.User, error) {
	row := d.Pool.QueryRow(ctx, `
		INSERT INTO users (email, firebase_uid)
		VALUES ($1, $2)
		ON CONFLICT (firebase_uid) DO UPDATE SET email = EXCLUDED.email
		RETURNING `+userColumns, email, firebaseUID)
	return scanUser(row)
}

// GetUser loads a user by id.
func (d *DB) GetUser(ctx context.Context, id int64) (model.User, error) {
	row := d.Pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id)
	return scanUser(row)
}

// IncrementSessionEpoch invalidates all currently-issued session tokens for a
// user (used on logout) and returns the new epoch.
func (d *DB) IncrementSessionEpoch(ctx context.Context, id int64) (int, error) {
	var epoch int
	err := d.Pool.QueryRow(ctx,
		`UPDATE users SET session_epoch = session_epoch + 1 WHERE id = $1 RETURNING session_epoch`,
		id).Scan(&epoch)
	return epoch, err
}

// UpdateUserProfile sets the display name and timezone.
func (d *DB) UpdateUserProfile(ctx context.Context, id int64, displayName, timezone string) error {
	_, err := d.Pool.Exec(ctx,
		`UPDATE users SET display_name = $2, timezone = $3 WHERE id = $1`,
		id, displayName, timezone)
	return err
}

// ListUsers returns all participants.
func (d *DB) ListUsers(ctx context.Context) ([]model.User, error) {
	rows, err := d.Pool.Query(ctx, `SELECT `+userColumns+` FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
