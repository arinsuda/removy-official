package migrate

// BackfillOnboardedAt marks every existing account as onboarded so production users are not
// pushed through onboarding again. Accounts created later start with NULL.
const BackfillOnboardedAt = `
UPDATE users SET onboarded_at = COALESCE(first_login_at, created_at)
WHERE onboarded_at IS NULL AND deleted_at IS NULL;
`
