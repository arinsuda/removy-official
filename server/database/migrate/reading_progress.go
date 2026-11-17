package migrate

import (
	"fmt"

	"github.com/arinsuda/movie-hub/internal/shared"
)

// ReadingProgressChecks adds the CHECK constraints AutoMigrate cannot express for reading_progress: manga
// only, chapter and volume inside their caps, and at least one of them above zero. The caps come from
// config so the app and the database agree. The table is new, so the constraints are added without NOT VALID
// and the table is scanned once, empty. The caller records the version in migration_history.
func ReadingProgressChecks(maxChapter, maxVolume int) string {
	return fmt.Sprintf(`
ALTER TABLE reading_progress DROP CONSTRAINT IF EXISTS chk_reading_progress_media_type;
ALTER TABLE reading_progress ADD CONSTRAINT chk_reading_progress_media_type
	CHECK (media_type = '%s');
ALTER TABLE reading_progress DROP CONSTRAINT IF EXISTS chk_reading_progress_range;
ALTER TABLE reading_progress ADD CONSTRAINT chk_reading_progress_range
	CHECK (chapter BETWEEN 0 AND %d AND volume BETWEEN 0 AND %d AND (chapter > 0 OR volume > 0));
`, shared.MediaTypeManga, maxChapter, maxVolume)
}
