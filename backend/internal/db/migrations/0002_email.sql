-- Switch the user identity from phone to email (passwordless email-link login).
-- Phone becomes optional; email is the contact identifier going forward.
ALTER TABLE users ADD COLUMN IF NOT EXISTS email TEXT;
ALTER TABLE users ALTER COLUMN phone DROP NOT NULL;
