package event_module

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/arinsuda/movie-hub/internal/admin_module"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	defaultPageLimit = 20
	maxPageLimit     = 50
	// popupLimit caps how many events the login popup cycles through.
	popupLimit = 5
	// MaxImageBytes bounds an uploaded banner. main.go derives the request body limit from it.
	MaxImageBytes  = 5 << 20
	eventImageRoot = "events/"
	auditTarget    = "event"
)

var (
	ErrNotFound      = errors.New("event not found")
	ErrImageTooLarge = errors.New("image too large")
)

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

func extFor(contentType string) string {
	if ext, ok := allowedImageTypes[contentType]; ok {
		return ext
	}
	return ".jpg"
}

// ImageStore is the slice of storage.MinIOClient the module needs.
type ImageStore interface {
	UploadBuffer(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (string, error)
	PresignURL(ctx context.Context, objectKey string) (string, error)
	DeleteObject(ctx context.Context, objectKey string) error
}

type Service struct {
	db    *gorm.DB
	store ImageStore
	now   func() time.Time
}

func NewService(db *gorm.DB, store ImageStore) *Service {
	return &Service{db: db, store: store, now: time.Now}
}

func clampPage(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultPageLimit
	}
	if limit > maxPageLimit {
		limit = maxPageLimit
	}
	return page, limit
}

// publishedScope limits a query to events anyone may see.
func publishedScope(db *gorm.DB) *gorm.DB {
	return db.Where("status = ?", StatusPublished)
}

func stateScope(state EventState, now time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		switch state {
		case StateUpcoming:
			return db.Where("starts_at > ?", now)
		case StatePast:
			return db.Where("ends_at <= ?", now)
		default:
			return db.Where("starts_at <= ? AND ends_at > ?", now, now)
		}
	}
}

func stateOrder(state EventState) string {
	switch state {
	case StateUpcoming:
		return "starts_at ASC, id ASC"
	case StatePast:
		return "ends_at DESC, id DESC"
	default:
		return "starts_at DESC, id DESC"
	}
}

func (s *Service) ListPublic(ctx context.Context, state EventState, page, limit int) ([]EventResponse, Pagination, error) {
	page, limit = clampPage(page, limit)
	now := s.now().UTC()
	q := s.db.WithContext(ctx).Model(&Event{}).Scopes(publishedScope, stateScope(state, now))

	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, Pagination{}, err
	}
	var rows []Event
	if err := q.Order(stateOrder(state)).Limit(limit).Offset((page - 1) * limit).Find(&rows).Error; err != nil {
		return nil, Pagination{}, err
	}
	out := make([]EventResponse, 0, len(rows))
	for i := range rows {
		out = append(out, s.toResponse(ctx, &rows[i], now))
	}
	return out, newPagination(page, limit, total), nil
}

func (s *Service) GetPublic(ctx context.Context, id uint) (EventResponse, error) {
	var ev Event
	err := s.db.WithContext(ctx).Scopes(publishedScope).First(&ev, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return EventResponse{}, ErrNotFound
	}
	if err != nil {
		return EventResponse{}, err
	}
	return s.toResponse(ctx, &ev, s.now().UTC()), nil
}

// Popup returns the active published events flagged for the login popup.
func (s *Service) Popup(ctx context.Context) ([]EventResponse, error) {
	now := s.now().UTC()
	var rows []Event
	err := s.db.WithContext(ctx).Scopes(publishedScope).
		Where("show_popup = ? AND starts_at <= ? AND ends_at > ?", true, now, now).
		Order("popup_priority DESC, starts_at DESC, id DESC").
		Limit(popupLimit).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]EventResponse, 0, len(rows))
	for i := range rows {
		out = append(out, s.toResponse(ctx, &rows[i], now))
	}
	return out, nil
}

func (s *Service) ListAdmin(ctx context.Context, status EventStatus, page, limit int) ([]AdminEventResponse, Pagination, error) {
	page, limit = clampPage(page, limit)
	q := s.db.WithContext(ctx).Model(&Event{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, Pagination{}, err
	}
	var rows []Event
	if err := q.Order("starts_at DESC, id DESC").Limit(limit).Offset((page - 1) * limit).Find(&rows).Error; err != nil {
		return nil, Pagination{}, err
	}
	now := s.now().UTC()
	out := make([]AdminEventResponse, 0, len(rows))
	for i := range rows {
		out = append(out, s.toAdminResponse(ctx, &rows[i], now))
	}
	return out, newPagination(page, limit, total), nil
}

func (s *Service) Create(ctx context.Context, adminID uint, in EventInput) (AdminEventResponse, error) {
	if err := in.normalize(); err != nil {
		return AdminEventResponse{}, err
	}
	ev := Event{Type: TypeAnnouncement, CreatedBy: &adminID}
	applyInput(&ev, in)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&ev).Error; err != nil {
			return err
		}
		return writeAudit(tx, adminID, admin_module.ActionEventCreated, ev.ID, map[string]any{"status": ev.Status})
	})
	if err != nil {
		return AdminEventResponse{}, err
	}
	return s.toAdminResponse(ctx, &ev, s.now().UTC()), nil
}

func (s *Service) Update(ctx context.Context, adminID, id uint, in EventInput) (AdminEventResponse, error) {
	if err := in.normalize(); err != nil {
		return AdminEventResponse{}, err
	}
	var ev Event
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockEvent(tx, id, &ev); err != nil {
			return err
		}
		prev := ev.Status
		applyInput(&ev, in)
		// Write only form columns: image_key and created_* are owned by other paths.
		if err := tx.Model(&ev).Select(eventFormColumns).Updates(&ev).Error; err != nil {
			return err
		}
		return writeAudit(tx, adminID, admin_module.ActionEventUpdated, ev.ID, map[string]any{"from_status": prev, "status": ev.Status})
	})
	if err != nil {
		return AdminEventResponse{}, err
	}
	return s.toAdminResponse(ctx, &ev, s.now().UTC()), nil
}

// Delete soft-deletes the event. Its banner object is kept so the row stays restorable.
func (s *Service) Delete(ctx context.Context, adminID, id uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ev Event
		if err := lockEvent(tx, id, &ev); err != nil {
			return err
		}
		if err := tx.Delete(&ev).Error; err != nil {
			return err
		}
		return writeAudit(tx, adminID, admin_module.ActionEventDeleted, ev.ID, nil)
	})
}

// SetImage uploads a new banner, swaps the key in the row, then removes the old object.
func (s *Service) SetImage(ctx context.Context, adminID, id uint, file io.ReadSeeker, size int64, contentType string) (AdminEventResponse, error) {
	if size > MaxImageBytes {
		return AdminEventResponse{}, ErrImageTooLarge
	}
	if _, err := s.find(ctx, id); err != nil {
		return AdminEventResponse{}, err
	}
	key := fmt.Sprintf("%s%d/%s%s", eventImageRoot, id, uuid.NewString(), extFor(contentType))
	if _, err := s.store.UploadBuffer(ctx, key, file, size, contentType); err != nil {
		return AdminEventResponse{}, err
	}
	ev, oldKey, err := s.swapImage(ctx, adminID, id, key)
	if err != nil {
		s.deleteObject(ctx, key)
		return AdminEventResponse{}, err
	}
	s.deleteObject(ctx, oldKey)
	return s.toAdminResponse(ctx, &ev, s.now().UTC()), nil
}

func (s *Service) RemoveImage(ctx context.Context, adminID, id uint) (AdminEventResponse, error) {
	ev, oldKey, err := s.swapImage(ctx, adminID, id, "")
	if err != nil {
		return AdminEventResponse{}, err
	}
	s.deleteObject(ctx, oldKey)
	return s.toAdminResponse(ctx, &ev, s.now().UTC()), nil
}

func (s *Service) swapImage(ctx context.Context, adminID, id uint, key string) (Event, string, error) {
	var ev Event
	var oldKey string
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockEvent(tx, id, &ev); err != nil {
			return err
		}
		oldKey = ev.ImageKey
		ev.ImageKey = key
		if err := tx.Model(&ev).Update("image_key", key).Error; err != nil {
			return err
		}
		return writeAudit(tx, adminID, admin_module.ActionEventUpdated, ev.ID, map[string]any{"image": key != ""})
	})
	return ev, oldKey, err
}

func (s *Service) deleteObject(ctx context.Context, key string) {
	if key == "" {
		return
	}
	if err := s.store.DeleteObject(ctx, key); err != nil {
		log.Printf("event_module: delete image object: %v", err)
	}
}

func (s *Service) find(ctx context.Context, id uint) (Event, error) {
	var ev Event
	err := s.db.WithContext(ctx).First(&ev, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ev, ErrNotFound
	}
	return ev, err
}

// eventFormColumns are the fields an admin edit may change (see applyInput).
var eventFormColumns = []string{
	"TitleEN", "TitleTH", "BodyEN", "BodyTH", "LinkURL",
	"StartsAt", "EndsAt", "ShowPopup", "PopupPriority", "Status", "UpdatedAt",
}

func lockEvent(tx *gorm.DB, id uint, ev *Event) error {
	err := tx.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).First(ev, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func applyInput(ev *Event, in EventInput) {
	ev.TitleEN, ev.TitleTH = in.TitleEN, in.TitleTH
	ev.BodyEN, ev.BodyTH = in.BodyEN, in.BodyTH
	ev.LinkURL = in.LinkURL
	ev.StartsAt, ev.EndsAt = in.StartsAt, in.EndsAt
	ev.ShowPopup, ev.PopupPriority = in.ShowPopup, in.PopupPriority
	ev.Status = in.Status
}

func writeAudit(tx *gorm.DB, adminID uint, action admin_module.AuditAction, eventID uint, meta map[string]any) error {
	row := admin_module.AdminAuditLog{
		AdminID:    &adminID,
		Action:     action,
		TargetType: auditTarget,
		TargetID:   eventID,
		CreatedAt:  time.Now().UTC(),
	}
	if meta != nil {
		js, err := json.Marshal(meta)
		if err != nil {
			return err
		}
		row.MetaData = datatypes.JSON(js)
	}
	return tx.Create(&row).Error
}

func optional(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func (s *Service) toResponse(ctx context.Context, ev *Event, now time.Time) EventResponse {
	var image *string
	if ev.ImageKey != "" {
		if u, err := s.store.PresignURL(ctx, ev.ImageKey); err == nil {
			image = optional(u)
		} else {
			log.Printf("event_module: presign image: %v", err)
		}
	}
	return EventResponse{
		ID:            ev.ID,
		Type:          ev.Type,
		MediaType:     ev.MediaType,
		Period:        ev.Period,
		Title:         LocalizedText{EN: ev.TitleEN, TH: ev.TitleTH},
		Body:          LocalizedText{EN: ev.BodyEN, TH: ev.BodyTH},
		ImageURL:      image,
		LinkURL:       optional(ev.LinkURL),
		StartsAt:      ev.StartsAt,
		EndsAt:        ev.EndsAt,
		ShowPopup:     ev.ShowPopup,
		PopupPriority: ev.PopupPriority,
		State:         ev.StateAt(now),
	}
}

func (s *Service) toAdminResponse(ctx context.Context, ev *Event, now time.Time) AdminEventResponse {
	return AdminEventResponse{
		EventResponse: s.toResponse(ctx, ev, now),
		Status:        ev.Status,
		CreatedAt:     ev.CreatedAt,
		UpdatedAt:     ev.UpdatedAt,
	}
}
