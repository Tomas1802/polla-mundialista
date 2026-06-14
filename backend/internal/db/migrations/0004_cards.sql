-- Players own 1..3 cards (cartones). Predictions belong to a card, not a user.
-- A user (email login) is linked to the player they are in the pool.

CREATE TABLE IF NOT EXISTS players (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS cards (
    id        BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    card_no   INTEGER NOT NULL DEFAULT 1,
    label     TEXT NOT NULL DEFAULT '',
    UNIQUE (player_id, card_no)
);

CREATE TABLE IF NOT EXISTS card_predictions (
    id             BIGSERIAL PRIMARY KEY,
    card_id        BIGINT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    match_id       BIGINT NOT NULL REFERENCES matches(id),
    home           INTEGER,
    away           INTEGER,
    penalty_winner TEXT,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (card_id, match_id)
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS player_id BIGINT REFERENCES players(id);
