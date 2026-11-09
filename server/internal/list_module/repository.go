package list_module

import (
	"errors"

	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/shared"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type repository struct{ db *gorm.DB }

func newRepository(db *gorm.DB) *repository { return &repository{db: db} }

// lockOwnedList takes a row lock on the caller's live list so adds, removals, reorders and
// edits on one list run one at a time. Missing, deleted or foreign lists are ErrNotFound.
func lockOwnedList(tx *gorm.DB, listID, userID uint) (*UserList, error) {
	var list UserList
	err := tx.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Where("id = ? AND user_id = ?", listID, userID).
		First(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &list, err
}

func (r *repository) findList(listID uint) (*UserList, error) {
	var list UserList
	err := r.db.First(&list, listID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &list, err
}

// createList serialises per user with an advisory lock so the cap cannot be raced past.
func (r *repository) createList(list *UserList) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?, ?)", listLockNS, int64(list.UserID)).Error; err != nil {
			return err
		}
		var n int64
		if err := tx.Model(&UserList{}).Where("user_id = ?", list.UserID).Count(&n).Error; err != nil {
			return err
		}
		if n >= MaxListsPerUser {
			return ErrListLimit
		}
		return tx.Create(list).Error
	})
}

func (r *repository) listsByUser(userID uint, visibilities []privacy_policy.ListVisibility) ([]UserList, error) {
	var lists []UserList
	q := r.db.Where("user_id = ?", userID)
	if visibilities != nil {
		q = q.Where("visibility IN ?", visibilities)
	}
	err := q.Order("updated_at DESC, id DESC").Find(&lists).Error
	return lists, err
}

func (r *repository) updateList(listID, userID uint, apply func(tx *gorm.DB, list *UserList) (map[string]any, error)) (*UserList, error) {
	var out *UserList
	err := r.db.Transaction(func(tx *gorm.DB) error {
		list, err := lockOwnedList(tx, listID, userID)
		if err != nil {
			return err
		}
		updates, err := apply(tx, list)
		if err != nil {
			return err
		}
		if err := tx.Model(list).Updates(updates).Error; err != nil {
			return err
		}
		out = list
		return nil
	})
	return out, err
}

func (r *repository) deleteList(listID, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		list, err := lockOwnedList(tx, listID, userID)
		if err != nil {
			return err
		}
		return tx.Delete(list).Error
	})
}

func (r *repository) addItem(listID, userID uint, mediaType shared.MediaType, mediaID int) (*UserListItem, error) {
	var item *UserListItem
	err := r.db.Transaction(func(tx *gorm.DB) error {
		list, err := lockOwnedList(tx, listID, userID)
		if err != nil {
			return err
		}
		var stats struct {
			N   int64
			Max int
		}
		if err := tx.Model(&UserListItem{}).Select("COUNT(*) AS n, COALESCE(MAX(position), 0) AS max").
			Where("list_id = ?", listID).Scan(&stats).Error; err != nil {
			return err
		}
		if stats.N >= MaxItemsPerList {
			return ErrListFull
		}
		it := &UserListItem{ListID: listID, MediaType: string(mediaType), MediaID: mediaID, Position: stats.Max + 1}
		if err := tx.Create(it).Error; err != nil {
			if shared.IsPgUniqueViolationOn(err, uniqueListMediaIndex) {
				return ErrItemExists
			}
			return err
		}
		item = it
		return tx.Model(list).UpdateColumn("updated_at", gorm.Expr("now()")).Error
	})
	return item, err
}

func (r *repository) removeItem(listID, userID, itemID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		list, err := lockOwnedList(tx, listID, userID)
		if err != nil {
			return err
		}
		res := tx.Where("id = ? AND list_id = ?", itemID, listID).Delete(&UserListItem{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrItemNotFound
		}
		updates := map[string]any{"updated_at": gorm.Expr("now()")}
		if list.CoverItemID != nil && *list.CoverItemID == itemID {
			updates["cover_item_id"] = nil
		}
		return tx.Model(list).UpdateColumns(updates).Error
	})
}

// reorder rewrites positions to 1..n in the requested order; itemIDs must be exactly the list's items.
func (r *repository) reorder(listID, userID uint, itemIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		list, err := lockOwnedList(tx, listID, userID)
		if err != nil {
			return err
		}
		var current []uint
		if err := tx.Model(&UserListItem{}).Where("list_id = ?", listID).Pluck("id", &current).Error; err != nil {
			return err
		}
		if !samePermutation(current, itemIDs) {
			return ErrOrderMismatch
		}
		for i, id := range itemIDs {
			if err := tx.Model(&UserListItem{}).Where("id = ? AND list_id = ?", id, listID).
				UpdateColumn("position", i+1).Error; err != nil {
				return err
			}
		}
		return tx.Model(list).UpdateColumn("updated_at", gorm.Expr("now()")).Error
	})
}

func (r *repository) itemsPage(listID uint, offset, limit int) ([]UserListItem, int64, error) {
	var total int64
	if err := r.db.Model(&UserListItem{}).Where("list_id = ?", listID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []UserListItem
	err := r.db.Where("list_id = ?", listID).Order("position ASC, id ASC").
		Offset(offset).Limit(limit).Find(&items).Error
	return items, total, err
}

// summaryData gathers counts, cover items, collage items and (optionally) media membership for lists.
type summaryData struct {
	counts   map[uint]int
	covers   map[uint]UserListItem // by list id
	collages map[uint][]UserListItem
	contains map[uint]bool
}

func (r *repository) summaryData(lists []UserList, mediaType string, mediaID int) (*summaryData, error) {
	d := &summaryData{counts: map[uint]int{}, covers: map[uint]UserListItem{}, collages: map[uint][]UserListItem{}, contains: map[uint]bool{}}
	if len(lists) == 0 {
		return d, nil
	}
	ids := make([]uint, len(lists))
	var coverIDs []uint
	for i, l := range lists {
		ids[i] = l.ID
		if l.CoverItemID != nil {
			coverIDs = append(coverIDs, *l.CoverItemID)
		}
	}

	var counts []struct {
		ListID uint
		N      int
	}
	if err := r.db.Model(&UserListItem{}).Select("list_id, COUNT(*) AS n").
		Where("list_id IN ?", ids).Group("list_id").Scan(&counts).Error; err != nil {
		return nil, err
	}
	for _, c := range counts {
		d.counts[c.ListID] = c.N
	}

	if len(coverIDs) > 0 {
		var covers []UserListItem
		if err := r.db.Where("id IN ? AND list_id IN ?", coverIDs, ids).Find(&covers).Error; err != nil {
			return nil, err
		}
		for _, c := range covers {
			d.covers[c.ListID] = c
		}
	}

	var collage []UserListItem
	err := r.db.Raw(`SELECT id, list_id, media_type, media_id, position, created_at FROM (
		SELECT i.*, row_number() OVER (PARTITION BY list_id ORDER BY position, id) AS rn
		FROM user_list_items i WHERE list_id IN ?) t WHERE rn <= ? ORDER BY list_id, position, id`, ids, CollagePosterCount).
		Scan(&collage).Error
	if err != nil {
		return nil, err
	}
	for _, it := range collage {
		d.collages[it.ListID] = append(d.collages[it.ListID], it)
	}

	if mediaID > 0 && mediaType != "" {
		var hit []uint
		if err := r.db.Model(&UserListItem{}).
			Where("list_id IN ? AND media_type = ? AND media_id = ?", ids, mediaType, mediaID).
			Pluck("list_id", &hit).Error; err != nil {
			return nil, err
		}
		for _, id := range hit {
			d.contains[id] = true
		}
	}
	return d, nil
}
