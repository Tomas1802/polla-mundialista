-- App settings (single row). edit_lock_until_match_id: matches up to and
-- including this one (in chronological order) are NOT editable by players
-- ("Temporalmente no editable"). NULL = no lock. Set from the admin panel.
CREATE TABLE IF NOT EXISTS app_settings (
    id                       INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    edit_lock_until_match_id BIGINT REFERENCES matches(id)
);
INSERT INTO app_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;
