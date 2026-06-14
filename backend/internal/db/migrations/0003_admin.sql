-- Admin role and manually-entered (admin-authoritative) match results.
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;

-- When result_manual is true, the football-data sync must NOT overwrite the
-- score/winner/status: the admin's entry is authoritative.
ALTER TABLE matches ADD COLUMN IF NOT EXISTS result_manual BOOLEAN NOT NULL DEFAULT FALSE;
