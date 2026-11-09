package migrate

// EmailVerificationTokenHash replaces stored raw verification tokens with their
// SHA-256 hex digest (what auth_module.HashToken produces). Expired rows are dropped
// first; unexpired links already mailed keep working because lookup hashes the
// presented token.
const EmailVerificationTokenHash = `
DELETE FROM email_verifications WHERE expires_at < now();
UPDATE email_verifications SET token = encode(sha256(convert_to(token, 'UTF8')), 'hex');
`
