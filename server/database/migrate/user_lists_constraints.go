package migrate

// UserListsConstraints adds the CHECK constraints AutoMigrate cannot express for
// user_lists / user_list_items. Guarded so a rerun or a partially applied schema is safe.
const UserListsConstraints = `
DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_user_lists_visibility') THEN
		ALTER TABLE user_lists ADD CONSTRAINT chk_user_lists_visibility
			CHECK (visibility IN ('public', 'followers', 'private'));
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_user_list_items_media_type') THEN
		ALTER TABLE user_list_items ADD CONSTRAINT chk_user_list_items_media_type
			CHECK (media_type IN ('movie', 'tv'));
	END IF;
END $$;
`
