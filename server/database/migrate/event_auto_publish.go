package migrate

import "gorm.io/gorm"

// BackfillEventAutoPublish keeps pre-existing worker drafts publishing: before the auto_publish
// column, auto_generated drafts were published by the worker, so they are opted in once. Rounds that
// are already published, cancelled, deleted or admin-made keep false.
func BackfillEventAutoPublish(tx *gorm.DB) error {
	return tx.Exec(`UPDATE events SET auto_publish = true
		WHERE auto_generated AND status = 'draft' AND deleted_at IS NULL`).Error
}
