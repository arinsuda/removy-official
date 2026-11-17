package config

// Reading progress (manga chapter and volume a signed-in user reached) is stored per user and title.
// The caps are mirrored by MANGA_PROGRESS_MAX_CHAPTER and MANGA_PROGRESS_MAX_VOLUME in
// client/src/components/manga/mangaConfig.ts; change both together.
const (
	// ReadingProgressMaxChapter is the highest chapter number a user may record. Longest-running manga
	// are well under this; the cap keeps a typo from becoming a real-looking value.
	ReadingProgressMaxChapter = 9999
	// ReadingProgressMaxVolume is the highest volume number a user may record.
	ReadingProgressMaxVolume = 999
)
