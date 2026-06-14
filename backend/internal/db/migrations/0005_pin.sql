-- Player PIN authentication (replaces email/Firebase login).
-- pin_hash: bcrypt of the current PIN. initial_pin: the assigned plaintext PIN,
-- visible to the admin for distribution until the player changes it.
ALTER TABLE players ADD COLUMN IF NOT EXISTS pin_hash        TEXT;
ALTER TABLE players ADD COLUMN IF NOT EXISTS initial_pin     TEXT;
ALTER TABLE players ADD COLUMN IF NOT EXISTS must_change_pin BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE players ADD COLUMN IF NOT EXISTS session_epoch   INTEGER NOT NULL DEFAULT 0;
