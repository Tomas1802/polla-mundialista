-- Pool teams (equipos): up to 3 players each. A team's points = the sum of each
-- member's best card. Players with no team (pool_team_id NULL) are "sin equipo".
CREATE TABLE IF NOT EXISTS pool_teams (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

ALTER TABLE players ADD COLUMN IF NOT EXISTS pool_team_id BIGINT REFERENCES pool_teams(id);
