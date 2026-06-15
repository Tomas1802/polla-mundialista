package db

import (
	"context"

	"polla/internal/model"
)

// UpsertPlayer inserts a player by name (or returns the existing id).
func (d *DB) UpsertPlayer(ctx context.Context, name string) (int64, error) {
	var id int64
	err := d.Pool.QueryRow(ctx, `
		INSERT INTO players (name) VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, name).Scan(&id)
	return id, err
}

// UpsertCard inserts or updates a card for a player.
func (d *DB) UpsertCard(ctx context.Context, playerID int64, cardNo int, label string) (int64, error) {
	var id int64
	err := d.Pool.QueryRow(ctx, `
		INSERT INTO cards (player_id, card_no, label) VALUES ($1, $2, $3)
		ON CONFLICT (player_id, card_no) DO UPDATE SET label = EXCLUDED.label
		RETURNING id`, playerID, cardNo, label).Scan(&id)
	return id, err
}

// ListPlayers returns all players with their card counts, ordered by name.
func (d *DB) ListPlayers(ctx context.Context) ([]model.Player, error) {
	rows, err := d.Pool.Query(ctx, `
		SELECT p.id, p.name, COUNT(c.id)
		FROM players p
		LEFT JOIN cards c ON c.player_id = p.id
		GROUP BY p.id, p.name
		ORDER BY p.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Player
	for rows.Next() {
		var p model.Player
		if err := rows.Scan(&p.ID, &p.Name, &p.CardCount); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetPlayerName returns a player's name.
func (d *DB) GetPlayerName(ctx context.Context, id int64) (string, error) {
	var name string
	err := d.Pool.QueryRow(ctx, `SELECT name FROM players WHERE id = $1`, id).Scan(&name)
	return name, err
}

// GetPlayer returns a player's auth-relevant fields.
func (d *DB) GetPlayer(ctx context.Context, id int64) (model.Player, error) {
	var p model.Player
	err := d.Pool.QueryRow(ctx, `
		SELECT id, name, COALESCE(pin_hash, ''), must_change_pin, session_epoch
		FROM players WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.PinHash, &p.MustChangePin, &p.SessionEpoch)
	return p, err
}

// SetPlayerPin sets a player's PIN hash and forces a change on next login.
// Used by the PIN generator (cmd/genpins); the plaintext lives only in the CSV.
func (d *DB) SetPlayerPin(ctx context.Context, id int64, hash string) error {
	_, err := d.Pool.Exec(ctx, `
		UPDATE players SET pin_hash = $2, initial_pin = NULL, must_change_pin = TRUE
		WHERE id = $1`, id, hash)
	return err
}

// ChangePin replaces a player's PIN and clears the initial one.
func (d *DB) ChangePin(ctx context.Context, id int64, hash string) error {
	_, err := d.Pool.Exec(ctx, `
		UPDATE players SET pin_hash = $2, initial_pin = NULL, must_change_pin = FALSE
		WHERE id = $1`, id, hash)
	return err
}

// IncrementPlayerEpoch invalidates a player's existing sessions (logout).
func (d *DB) IncrementPlayerEpoch(ctx context.Context, id int64) (int, error) {
	var epoch int
	err := d.Pool.QueryRow(ctx,
		`UPDATE players SET session_epoch = session_epoch + 1 WHERE id = $1 RETURNING session_epoch`,
		id).Scan(&epoch)
	return epoch, err
}

// ListPlayersAdmin returns all players with card counts and whether they still
// have their assigned PIN (must_change_pin) — without exposing any PIN value.
func (d *DB) ListPlayersAdmin(ctx context.Context) ([]model.Player, error) {
	rows, err := d.Pool.Query(ctx, `
		SELECT p.id, p.name, COUNT(c.id), p.must_change_pin, (p.pin_hash IS NOT NULL)
		FROM players p
		LEFT JOIN cards c ON c.player_id = p.id
		GROUP BY p.id, p.name, p.must_change_pin, p.pin_hash
		ORDER BY p.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Player
	for rows.Next() {
		var p model.Player
		var hasPin bool
		if err := rows.Scan(&p.ID, &p.Name, &p.CardCount, &p.MustChangePin, &hasPin); err != nil {
			return nil, err
		}
		p.HasPin = hasPin
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListCardsByPlayer returns a player's cards ordered by card number.
func (d *DB) ListCardsByPlayer(ctx context.Context, playerID int64) ([]model.Card, error) {
	rows, err := d.Pool.Query(ctx, `
		SELECT id, player_id, card_no, label FROM cards
		WHERE player_id = $1 ORDER BY card_no`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCards(rows)
}

// ListAllCards returns every card.
func (d *DB) ListAllCards(ctx context.Context) ([]model.Card, error) {
	rows, err := d.Pool.Query(ctx, `SELECT id, player_id, card_no, label FROM cards`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCards(rows)
}

// CardWithPlayer is a card plus its owner's name (for the admin card picker).
type CardWithPlayer struct {
	CardID     int64  `json:"id"`
	CardNo     int    `json:"cardNo"`
	Label      string `json:"label"`
	PlayerName string `json:"playerName"`
}

// ListCardsWithPlayer returns every card with its player's name, ordered.
func (d *DB) ListCardsWithPlayer(ctx context.Context) ([]CardWithPlayer, error) {
	rows, err := d.Pool.Query(ctx, `
		SELECT c.id, c.card_no, c.label, p.name
		FROM cards c JOIN players p ON p.id = c.player_id
		ORDER BY p.name, c.card_no`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CardWithPlayer
	for rows.Next() {
		var c CardWithPlayer
		if err := rows.Scan(&c.CardID, &c.CardNo, &c.Label, &c.PlayerName); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// GetCard returns a single card (for ownership checks).
func (d *DB) GetCard(ctx context.Context, id int64) (model.Card, error) {
	var c model.Card
	err := d.Pool.QueryRow(ctx, `SELECT id, player_id, card_no, label FROM cards WHERE id = $1`, id).
		Scan(&c.ID, &c.PlayerID, &c.CardNo, &c.Label)
	return c, err
}

func scanCards(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]model.Card, error) {
	var out []model.Card
	for rows.Next() {
		var c model.Card
		if err := rows.Scan(&c.ID, &c.PlayerID, &c.CardNo, &c.Label); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// UpsertCardPrediction stores or updates a card's marcador for a match.
func (d *DB) UpsertCardPrediction(ctx context.Context, p model.CardPrediction) error {
	_, err := d.Pool.Exec(ctx, `
		INSERT INTO card_predictions (card_id, match_id, home, away, penalty_winner, updated_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (card_id, match_id) DO UPDATE SET
			home = EXCLUDED.home,
			away = EXCLUDED.away,
			penalty_winner = EXCLUDED.penalty_winner,
			updated_at = now()`,
		p.CardID, p.MatchID, p.Home, p.Away, nullStr(p.PenaltyWinner))
	return err
}

// DeleteCardPredictionsForMatches removes all cards' predictions for the given
// matches. Used at import start so a re-import cleanly replaces group-stage
// predictions instead of leaving stale rows from a previous (wrong) mapping.
func (d *DB) DeleteCardPredictionsForMatches(ctx context.Context, matchIDs []int64) error {
	if len(matchIDs) == 0 {
		return nil
	}
	_, err := d.Pool.Exec(ctx, `DELETE FROM card_predictions WHERE match_id = ANY($1)`, matchIDs)
	return err
}

// ListCardPredictions returns one card's predictions.
func (d *DB) ListCardPredictions(ctx context.Context, cardID int64) ([]model.CardPrediction, error) {
	rows, err := d.Pool.Query(ctx, `
		SELECT card_id, match_id, home, away, COALESCE(penalty_winner, '')
		FROM card_predictions WHERE card_id = $1`, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCardPredictions(rows)
}

// ListAllCardPredictions returns every card's predictions grouped by card id.
func (d *DB) ListAllCardPredictions(ctx context.Context) (map[int64][]model.CardPrediction, error) {
	rows, err := d.Pool.Query(ctx, `
		SELECT card_id, match_id, home, away, COALESCE(penalty_winner, '')
		FROM card_predictions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	all, err := scanCardPredictions(rows)
	if err != nil {
		return nil, err
	}
	byCard := map[int64][]model.CardPrediction{}
	for _, p := range all {
		byCard[p.CardID] = append(byCard[p.CardID], p)
	}
	return byCard, nil
}

func scanCardPredictions(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]model.CardPrediction, error) {
	var out []model.CardPrediction
	for rows.Next() {
		var p model.CardPrediction
		if err := rows.Scan(&p.CardID, &p.MatchID, &p.Home, &p.Away, &p.PenaltyWinner); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
