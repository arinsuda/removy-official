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
	ErrNotFound         = errors.New("event not found")
	ErrImageTooLarge    = errors.New("image too large")
	ErrVotingClosed     = errors.New("voting closed")
	ErrVotingNotStarted = errors.New("voting not started")
	ErrInvalidEventType = errors.New("invalid event type")
	ErrNomineeNotFound  = errors.New("nominee not found")
	ErrEventExists      = errors.New("event already exists")
	// ErrVotingLocked: the round has opened, so its type, window and nominees are frozen.
	ErrVotingLocked = errors.New("voting round is locked")
	// ErrVotingNotEnded: finalization was requested before ends_at.
	ErrVotingNotEnded = errors.New("voting has not ended")
	// ErrNotEnoughNominees: suggestions could not fill a round with real titles.
	ErrNotEnoughNominees = errors.New("not enough nominee candidates")
	// ErrVotingNotSettled: an ended round is still unfinalized after settlement (should not happen).
	ErrVotingNotSettled = errors.New("voting round not settled")
	// ErrEventNotPublished: only a published round can be finalized (drafts and cancelled rounds cannot).
	ErrEventNotPublished = errors.New("event is not published")
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
	covers, err := s.coverBackdrops(ctx, rows)
	if err != nil {
		return nil, Pagination{}, err
	}
	out := make([]EventResponse, 0, len(rows))
	for i := range rows {
		resp := s.toResponse(ctx, &rows[i], now)
		resp.CoverBackdropPath = optional(covers[rows[i].ID])
		out = append(out, resp)
	}
	return out, newPagination(page, limit, total), nil
}

func (s *Service) GetPublic(ctx context.Context, id uint, viewerID uint) (EventResponse, error) {
	var ev Event
	err := s.db.WithContext(ctx).Scopes(publishedScope).First(&ev, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return EventResponse{}, ErrNotFound
	}
	if err != nil {
		return EventResponse{}, err
	}
	now := s.now().UTC()

	// Lazy finalization: the first read after ends_at settles the round.
	if ev.Type.IsVoting() && ev.StateAt(now) == StatePast && !ev.IsFinalized {
		if err := s.finalizeTx(ctx, 0, ev.ID); err != nil {
			return EventResponse{}, err
		}
		if err := s.db.WithContext(ctx).First(&ev, id).Error; err != nil {
			return EventResponse{}, err
		}
	}

	resp := s.toResponse(ctx, &ev, now)
	covers, err := s.coverBackdrops(ctx, []Event{ev})
	if err != nil {
		return EventResponse{}, err
	}
	resp.CoverBackdropPath = optional(covers[ev.ID])
	if !ev.Type.IsVoting() {
		return resp, nil
	}

	nominees, err := s.loadNominees(ctx, ev.ID)
	if err != nil {
		return EventResponse{}, err
	}
	if viewerID > 0 {
		var myVote EventVote
		err := s.db.WithContext(ctx).Where("event_id = ? AND user_id = ?", id, viewerID).First(&myVote).Error
		switch {
		case err == nil:
			resp.MyVoteNomineeID = &myVote.NomineeID
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return EventResponse{}, err
		}
	}

	// Secrecy: counts and winners exist in the response only once voting has ended.
	if ev.StateAt(now) != StatePast {
		resp.Nominees = make([]NomineeResponse, 0, len(nominees))
		for i := range nominees {
			resp.Nominees = append(resp.Nominees, toNomineeResponse(&nominees[i]))
		}
		return resp, nil
	}

	// Results come from the settled snapshot, never from live votes.
	if !ev.IsFinalized {
		return EventResponse{}, ErrVotingNotSettled
	}
	totalVotes := 0
	for i := range nominees {
		totalVotes += finalVotes(nominees[i])
	}
	var winnerIDs []uint
	if err := s.db.WithContext(ctx).Model(&EventAward{}).Where("event_id = ?", id).Pluck("nominee_id", &winnerIDs).Error; err != nil {
		return EventResponse{}, err
	}
	winners := make(map[uint]bool, len(winnerIDs))
	for _, wid := range winnerIDs {
		winners[wid] = true
	}

	resp.TotalVotes = &totalVotes
	resp.Nominees = make([]NomineeResponse, 0, len(nominees))
	for i := range nominees {
		n := toNomineeResponse(&nominees[i])
		count := finalVotes(nominees[i])
		pct := votePercentage(count, totalVotes)
		n.VoteCount, n.Percentage, n.IsWinner = &count, &pct, winners[nominees[i].ID]
		resp.Nominees = append(resp.Nominees, n)
	}
	return resp, nil
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
	covers, err := s.coverBackdrops(ctx, rows)
	if err != nil {
		return nil, err
	}
	out := make([]EventResponse, 0, len(rows))
	for i := range rows {
		resp := s.toResponse(ctx, &rows[i], now)
		resp.CoverBackdropPath = optional(covers[rows[i].ID])
		out = append(out, resp)
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
	ids := make([]uint, 0, len(rows))
	for i := range rows {
		ids = append(ids, rows[i].ID)
	}
	byEvent, err := s.loadNomineesFor(ctx, ids)
	if err != nil {
		return nil, Pagination{}, err
	}
	now := s.now().UTC()
	out := make([]AdminEventResponse, 0, len(rows))
	for i := range rows {
		out = append(out, s.toAdminResponse(ctx, &rows[i], byEvent[rows[i].ID], now))
	}
	return out, newPagination(page, limit, total), nil
}

func (s *Service) Create(ctx context.Context, adminID uint, in EventInput) (AdminEventResponse, error) {
	if err := in.normalize(); err != nil {
		return AdminEventResponse{}, err
	}
	var creator *uint
	if adminID > 0 {
		creator = &adminID
	}
	ev := Event{CreatedBy: creator}
	applyInput(&ev, in)
	now := s.now().UTC()
	var nominees []EventNominee
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&ev).Error; err != nil {
			return mapEventUniqueErr(err)
		}
		var err error
		if nominees, err = replaceNominees(tx, ev.ID, in.Nominees, now); err != nil {
			return err
		}
		if err := checkPublishable(&ev, len(nominees)); err != nil {
			return err
		}
		if adminID > 0 {
			return writeAudit(tx, adminID, admin_module.ActionEventCreated, ev.ID, map[string]any{"status": ev.Status, "type": ev.Type})
		}
		return nil
	})
	if err != nil {
		return AdminEventResponse{}, err
	}
	return s.toAdminResponse(ctx, &ev, nominees, now), nil
}

func (s *Service) Update(ctx context.Context, adminID, id uint, in EventInput) (AdminEventResponse, error) {
	if err := in.normalize(); err != nil {
		return AdminEventResponse{}, err
	}
	var ev Event
	var nominees []EventNominee
	var now time.Time
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockEvent(tx, id, &ev); err != nil {
			return err
		}
		// Read the clock after the row lock, so a round that opened while we waited is refused.
		now = s.now().UTC()
		if err := checkVotingLock(&ev, in, now); err != nil {
			return err
		}
		prev := ev.Status
		prevType, prevMediaType := ev.Type, ev.MediaType
		applyInput(&ev, in)
		if err := tx.Model(&ev).Select(eventFormColumns).Updates(&ev).Error; err != nil {
			return mapEventUniqueErr(err)
		}
		var err error
		switch {
		case in.Nominees != nil:
			if err := checkNoVotes(tx, id); err != nil {
				return err
			}
			nominees, err = replaceNominees(tx, id, in.Nominees, now)
		case in.Type != prevType || !equalStringPtr(in.MediaType, prevMediaType):
			nominees, err = revalidateNominees(tx, id, in, now)
		default:
			nominees, err = loadNomineesTx(tx, id)
		}
		if err != nil {
			return err
		}
		// Re-check only when the publish state or the nominee set changes, so unrelated edits to a
		// published round (copy, window) are not blocked by a legacy round that has fewer titles.
		if prev != StatusPublished || in.Nominees != nil || in.Type != prevType || !equalStringPtr(in.MediaType, prevMediaType) {
			if err := checkPublishable(&ev, len(nominees)); err != nil {
				return err
			}
		}
		return writeAudit(tx, adminID, admin_module.ActionEventUpdated, ev.ID, map[string]any{"from_status": prev, "status": ev.Status})
	})
	if err != nil {
		return AdminEventResponse{}, err
	}
	return s.toAdminResponse(ctx, &ev, nominees, now), nil
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
	return s.adminResponseFor(ctx, &ev)
}

func (s *Service) RemoveImage(ctx context.Context, adminID, id uint) (AdminEventResponse, error) {
	ev, oldKey, err := s.swapImage(ctx, adminID, id, "")
	if err != nil {
		return AdminEventResponse{}, err
	}
	s.deleteObject(ctx, oldKey)
	return s.adminResponseFor(ctx, &ev)
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
	"Type", "MediaType", "Period",
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
	ev.Type = in.Type
	ev.MediaType = in.MediaType
	ev.Period = in.Period
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
		IsFinalized:   ev.IsFinalized,
	}
}

// coverBackdrops returns, for each voting round in rows without a banner, the backdrop of its first
// nominee (by position) that has one. Rounds with a banner or no backdrop are absent. One query.
func (s *Service) coverBackdrops(ctx context.Context, rows []Event) (map[uint]string, error) {
	ids := make([]uint, 0, len(rows))
	for i := range rows {
		if rows[i].Type.IsVoting() && rows[i].ImageKey == "" {
			ids = append(ids, rows[i].ID)
		}
	}
	out := make(map[uint]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	var covers []struct {
		EventID      uint
		BackdropPath string
	}
	err := s.db.WithContext(ctx).Model(&EventNominee{}).
		Select("DISTINCT ON (event_id) event_id, backdrop_path").
		Where("event_id IN ? AND backdrop_path <> ''", ids).
		Order("event_id, position, id").
		Scan(&covers).Error
	if err != nil {
		return nil, err
	}
	for _, c := range covers {
		out[c.EventID] = c.BackdropPath
	}
	return out, nil
}

func (s *Service) toAdminResponse(ctx context.Context, ev *Event, nominees []EventNominee, now time.Time) AdminEventResponse {
	resp := s.toResponse(ctx, ev, now)
	resp.Nominees = make([]NomineeResponse, 0, len(nominees))
	for i := range nominees {
		resp.Nominees = append(resp.Nominees, toNomineeResponse(&nominees[i]))
	}
	return AdminEventResponse{
		EventResponse: resp,
		Status:        ev.Status,
		CreatedAt:     ev.CreatedAt,
		UpdatedAt:     ev.UpdatedAt,
	}
}

// adminResponseFor builds the admin shape for one event, loading its nominees.
func (s *Service) adminResponseFor(ctx context.Context, ev *Event) (AdminEventResponse, error) {
	nominees, err := s.loadNominees(ctx, ev.ID)
	if err != nil {
		return AdminEventResponse{}, err
	}
	return s.toAdminResponse(ctx, ev, nominees, s.now().UTC()), nil
}
