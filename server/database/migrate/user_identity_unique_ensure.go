package migrate

// UserIdentityUniqueEnsure makes sure the identity uniqueness guarantees exist.
// AutoMigrate creates user_identities from the model first, which carries no
// unique constraints, so the raw-SQL definition in google_oauth_user_identity_v1
// is skipped on a fresh database. Additive only: an index is created only when
// absent and no duplicate rows exist; nothing is deleted or rewritten.
const UserIdentityUniqueEnsure = `
DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'uq_provider_subject') THEN
		IF EXISTS (SELECT 1 FROM user_identities GROUP BY provider, provider_subject HAVING count(*) > 1) THEN
			RAISE WARNING 'uq_provider_subject skipped: duplicate (provider, provider_subject) rows exist';
		ELSE
			CREATE UNIQUE INDEX uq_provider_subject ON user_identities (provider, provider_subject);
		END IF;
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'uq_user_provider') THEN
		IF EXISTS (SELECT 1 FROM user_identities GROUP BY user_id, provider HAVING count(*) > 1) THEN
			RAISE WARNING 'uq_user_provider skipped: duplicate (user_id, provider) rows exist';
		ELSE
			CREATE UNIQUE INDEX uq_user_provider ON user_identities (user_id, provider);
		END IF;
	END IF;
END $$;
`
