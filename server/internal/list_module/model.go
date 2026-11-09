package list_module

import (
	"time"

	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"gorm.io/gorm"
)

// Limits and sizes. Numbers here are the single source for validation, queries and tests.
const (
	MaxListsPerUser   = 50
	MaxItemsPerList   = 500
	MaxListNameRunes  = 80
	MaxListDescRunes  = 500
	ListItemsPageSize = 40

	// CollagePosterCount is how many posters the cover collage uses when no cover item is picked.
	CollagePosterCount = 4

	// listLockNS namespaces pg_advisory_xact_lock(ns, user_id) so list creation per user
	// serialises without colliding with other advisory lock users.
	listLockNS = 0x4c495354 // "LIST"

	tmdbConcurrency = 6
)

const (
	tableLists     = "user_lists"
	tableListItems = "user_list_items"

	// uniqueListMediaIndex is the unique index that rejects a title added twice to one list.
	uniqueListMediaIndex = "idx_user_list_items_list_media"
)

// UserList is a user-made list of titles. Soft deleted (DeletedAt), items are hard deleted.
type UserList struct {
	gorm.Model
	UserID      uint                          `gorm:"not null;index"`
	Name        string                        `gorm:"type:varchar(80);not null"`
	Description string                        `gorm:"type:varchar(500);not null;default:''"`
	Visibility  privacy_policy.ListVisibility `gorm:"type:varchar(20);not null;default:'public'"`
	CoverItemID *uint
}

func (UserList) TableName() string { return tableLists }

type UserListItem struct {
	ID        uint     `gorm:"primarykey;autoIncrement"`
	ListID    uint     `gorm:"not null;uniqueIndex:idx_user_list_items_list_media,priority:1"`
	List      UserList `gorm:"foreignKey:ListID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	MediaType string   `gorm:"type:varchar(10);not null;uniqueIndex:idx_user_list_items_list_media,priority:2"`
	MediaID   int      `gorm:"not null;uniqueIndex:idx_user_list_items_list_media,priority:3"`
	Position  int      `gorm:"not null;default:0"`
	CreatedAt time.Time
}

func (UserListItem) TableName() string { return tableListItems }
