package migrate

// MediaTypeManga widens the user_list_items media_type CHECK to accept manga (AniList ids).
// The old constraint was CHECK (media_type IN ('movie', 'tv')), a strict subset of the new one, so every
// existing row already satisfies it and no VALIDATE scan is needed. The drop and the NOT VALID add run in
// one DO block; NOT VALID still enforces the check on every new or updated row. The caller records the
// version in migration_history, so it runs once.
const MediaTypeManga = `
DO $$
BEGIN
	ALTER TABLE user_list_items DROP CONSTRAINT IF EXISTS chk_user_list_items_media_type;
	ALTER TABLE user_list_items ADD CONSTRAINT chk_user_list_items_media_type
		CHECK (media_type IN ('movie', 'tv', 'manga')) NOT VALID;
END $$;
`
