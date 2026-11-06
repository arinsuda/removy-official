package migrate

// AccountSecurityColumns adds the session epoch and the email-bound token
// columns. Every column is nullable or has a constant default, so existing rows
// are untouched and an old binary keeps running.
const AccountSecurityColumns = `
ALTER TABLE users ADD COLUMN IF NOT EXISTS token_version INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS identity_reset_at TIMESTAMPTZ;
ALTER TABLE email_verifications ADD COLUMN IF NOT EXISTS email VARCHAR(100);
ALTER TABLE password_reset_tokens ADD COLUMN IF NOT EXISTS email VARCHAR(100);
ALTER TABLE link_pending ADD COLUMN IF NOT EXISTS code_email VARCHAR(255);
ALTER TABLE email_change_requests ADD COLUMN IF NOT EXISTS new_otp_hash VARCHAR(255);
`
