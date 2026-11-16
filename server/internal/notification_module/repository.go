package notification_module

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

var ErrNotificationNotFound = errors.New("notification not found")

type repository struct {
	db *gorm.DB
}

func newRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, n *Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *repository) CreateBatch(ctx context.Context, ns []Notification) error {
	if len(ns) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&ns).Error
}

// mangaSafeSQL keeps every notification except one about manga whose manga_meta row is adult or missing.
// It is shared.MangaSafeSQL over the params column; shared imports this package, so the predicate is repeated here.
const mangaSafeSQL = `(COALESCE(params->>'media_type', '') <> 'manga' OR (params->>'media_id')::int IN (SELECT media_id FROM manga_meta WHERE is_adult = false))`

// withMangaFilter applies mangaSafeSQL when the viewer may not see adult manga.
func withMangaFilter(db *gorm.DB, restrictManga bool) *gorm.DB {
	if restrictManga {
		return db.Where(mangaSafeSQL)
	}
	return db
}

func (r *repository) FindByUser(ctx context.Context, userID uint, q ListNotificationsQuery, restrictManga bool) ([]Notification, int64, error) {
	page, pageSize := normalizePagination(q.Page, q.PageSize)

	db := r.db.WithContext(ctx).Model(&Notification{}).Where("user_id = ?", userID)
	db = withMangaFilter(db, restrictManga)

	if q.Unread != nil && *q.Unread {
		db = db.Where("is_read = false")
	}
	if q.Type != "" {
		db = db.Where("type = ?", q.Type)
	}
	if q.Category != "" {
		db = db.Where("category = ?", q.Category) // NEW
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []Notification
	err := db.Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&rows).Error

	return rows, total, err
}

// NEW: unread count แยกรายหมวด สำหรับ badge บนแท็บ
func (r *repository) CountUnreadByCategory(ctx context.Context, userID uint, restrictManga bool) ([]UnreadByCategoryResponse, error) {
	var rows []UnreadByCategoryResponse
	db := withMangaFilter(r.db.WithContext(ctx).Model(&Notification{}), restrictManga)
	err := db.Select("category, COUNT(*) as count").
		Where("user_id = ? AND is_read = false", userID).
		Group("category").
		Scan(&rows).Error
	return rows, err
}

func (r *repository) CountUnread(ctx context.Context, userID uint, restrictManga bool) (int64, error) {
	var count int64
	db := withMangaFilter(r.db.WithContext(ctx).Model(&Notification{}), restrictManga)
	err := db.Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error
	return count, err
}

func (r *repository) MarkRead(ctx context.Context, userID uint, ids []uint) error {
	db := r.db.WithContext(ctx).Model(&Notification{}).
		Where("user_id = ? AND is_read = false", userID)

	if len(ids) > 0 {
		db = db.Where("id IN ?", ids)
	}

	return db.Updates(map[string]any{
		"is_read": true,
		"read_at": time.Now(),
	}).Error
}

func (r *repository) DeleteByUser(ctx context.Context, userID uint, ids []uint) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND id IN ?", userID, ids).
		Delete(&Notification{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

func (r *repository) DeleteAllByUser(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&Notification{}).Error
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}
