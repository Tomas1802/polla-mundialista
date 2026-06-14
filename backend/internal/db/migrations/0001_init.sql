-- Polla Mundialista 2026 — initial schema (PostgreSQL / Cloud SQL).
-- Applied automatically on backend startup; recorded in schema_migrations.

-- ─────────────────────────────────────────────────────────────────────────────
-- Users. Identity is the phone number (verified via Firebase Phone Auth).
-- session_epoch is bumped on logout so previously-issued JWTs stop validating,
-- forcing a fresh OTP on the next login.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    phone         TEXT NOT NULL UNIQUE,          -- E.164, e.g. +573001234567
    firebase_uid  TEXT NOT NULL UNIQUE,
    display_name  TEXT NOT NULL DEFAULT '',
    timezone      TEXT NOT NULL DEFAULT 'UTC',   -- IANA tz reported by the browser
    session_epoch INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ─────────────────────────────────────────────────────────────────────────────
-- National teams (ids mirror football-data.org team ids).
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS teams (
    id           BIGINT PRIMARY KEY,
    name         TEXT NOT NULL,
    short_name   TEXT NOT NULL DEFAULT '',
    tla          TEXT NOT NULL DEFAULT '',        -- three-letter abbreviation
    crest_url    TEXT NOT NULL DEFAULT '',
    group_letter TEXT                              -- 'A'..'L' or NULL
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Matches: cached fixtures + results from football-data.org.
-- Knockout matches may have placeholder team names before the teams are decided
-- (e.g. "Winner Group A"), so we keep both an optional team id and a name
-- snapshot. score_* / winner / duration are NULL until the match finishes.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS matches (
    id              BIGINT PRIMARY KEY,           -- football-data match id
    utc_date        TIMESTAMPTZ NOT NULL,
    stage           TEXT NOT NULL,                -- GROUP_STAGE, LAST_32, LAST_16, ...
    group_letter    TEXT,                         -- 'A'..'L' for group stage, else NULL
    matchday        INTEGER,
    seq             INTEGER NOT NULL DEFAULT 0,   -- chronological ordering helper
    home_team_id    BIGINT REFERENCES teams(id),
    away_team_id    BIGINT REFERENCES teams(id),
    home_team_name  TEXT NOT NULL DEFAULT '',
    away_team_name  TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'SCHEDULED',
    score_home      INTEGER,
    score_away      INTEGER,
    winner          TEXT,                         -- HOME_WIN | AWAY_WIN | DRAW
    duration        TEXT,                         -- REGULAR | EXTRA_TIME | PENALTY_SHOOTOUT
    last_synced_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_matches_utc_date ON matches (utc_date);

-- ─────────────────────────────────────────────────────────────────────────────
-- Match-score predictions (marcadores). penalty_winner is only used for
-- knockout matches the user predicted as a draw.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS predictions (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    match_id       BIGINT NOT NULL REFERENCES matches(id),
    home           INTEGER,
    away           INTEGER,
    penalty_winner TEXT,                          -- HOME | AWAY | NULL
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, match_id)
);

-- NOTE: there is no separate table for predicted group positions or for bracket
-- matchups. Each user's predicted standings (section 2) and predicted bracket
-- are DERIVED on the fly from their predicted scores in `predictions` using the
-- standings engine. This keeps the model simple: users only ever enter
-- marcadores.

-- Official final group standings (filled once a group is decided).
CREATE TABLE IF NOT EXISTS group_standings (
    group_letter  TEXT PRIMARY KEY,
    pos1_team_id  BIGINT REFERENCES teams(id),
    pos2_team_id  BIGINT REFERENCES teams(id),
    pos3_team_id  BIGINT REFERENCES teams(id),
    pos4_team_id  BIGINT REFERENCES teams(id),
    decided       BOOLEAN NOT NULL DEFAULT FALSE
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Single-row bookkeeping for the football-data sync policy: one sync per day,
-- plus a re-sync once the next match's kickoff has passed.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS api_sync_state (
    id                 INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    last_full_sync_at  TIMESTAMPTZ,
    next_match_utc     TIMESTAMPTZ,
    next_match_id      BIGINT,
    next_match_synced  BOOLEAN NOT NULL DEFAULT FALSE
);
INSERT INTO api_sync_state (id) VALUES (1) ON CONFLICT (id) DO NOTHING;
