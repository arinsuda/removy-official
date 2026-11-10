package list_module

import (
	"context"
	"errors"

	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/shared"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

var (
	ErrNotFound          = errors.New("list not found")
	ErrItemNotFound      = errors.New("list item not found")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInvalidVisibility = errors.New("invalid visibility")
	ErrInvalidMediaType  = errors.New("invalid media type, must be movie or tv")
	ErrListLimit         = errors.New("list limit reached")
	ErrListFull          = errors.New("list is full")
	ErrItemExists        = errors.New("title already in this list")
	ErrOrderMismatch     = errors.New("item_ids must be exactly the list's items")
)

type Service struct {
	repo   *repository
	db     *gorm.DB
	policy privacy_policy.UserAccessPolicy
}

func NewService(db *gorm.DB, policy privacy_policy.UserAccessPolicy) *Service {
	return &Service{repo: newRepository(db), db: db, policy: policy}
}

func (s *Service) Create(userID uint, req CreateListRequest) (*ListSummary, error) {
	name, err := validName(req.Name)
	if err != nil {
		return nil, err
	}
	desc, err := validDescription(req.Description)
	if err != nil {
		return nil, err
	}
	vis := req.Visibility
	if vis == "" {
		vis = privacy_policy.ListPublic
	}
	if err := validVisibility(vis); err != nil {
		return nil, err
	}
	list := &UserList{UserID: userID, Name: name, Description: desc, Visibility: vis}
	if err := s.repo.createList(list); err != nil {
		return nil, err
	}
	return s.summarize([]UserList{*list}, "", 0)
}

// ListOwn returns the caller's lists; with mediaType+mediaID each carries contains_media.
func (s *Service) ListOwn(userID uint, mediaType shared.MediaType, mediaID int) ([]ListSummary, error) {
	lists, err := s.repo.listsByUser(userID, nil)
	if err != nil {
		return nil, err
	}
	out, err := s.summarizeAll(lists, string(mediaType), mediaID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ListForUser returns the lists of ownerID that viewerID may see. A locked profile is
// ErrForbidden, an inactive or missing owner is ErrNotFound.
func (s *Service) ListForUser(ctx context.Context, viewerID, ownerID uint) ([]ListSummary, error) {
	if viewerID == ownerID {
		return s.ListOwn(ownerID, "", 0)
	}
	canPublic, err := s.policy.CanViewList(ctx, viewerID, ownerID, privacy_policy.ListPublic)
	if err != nil {
		return nil, mapPolicyErr(err)
	}
	if !canPublic {
		return nil, ErrForbidden
	}
	visible := []privacy_policy.ListVisibility{privacy_policy.ListPublic}
	canFollowers, err := s.policy.CanViewList(ctx, viewerID, ownerID, privacy_policy.ListFollowers)
	if err != nil {
		return nil, mapPolicyErr(err)
	}
	if canFollowers {
		visible = append(visible, privacy_policy.ListFollowers)
	}
	lists, err := s.repo.listsByUser(ownerID, visible)
	if err != nil {
		return nil, err
	}
	return s.summarizeAll(lists, "", 0)
}

// Get returns one page of a list. A list the viewer may not open is ErrNotFound, never a 403,
// so private lists do not reveal that they exist.
func (s *Service) Get(ctx context.Context, viewerID, listID uint, page int) (*ListDetailResponse, error) {
	list, err := s.repo.findList(listID)
	if err != nil {
		return nil, err
	}
	canView, err := s.policy.CanViewList(ctx, viewerID, list.UserID, list.Visibility)
	if err != nil {
		return nil, mapPolicyErr(err)
	}
	if !canView {
		return nil, ErrNotFound
	}
	if page < 1 {
		page = 1
	}
	rows, total, err := s.repo.itemsPage(listID, (page-1)*ListItemsPageSize, ListItemsPageSize)
	if err != nil {
		return nil, err
	}
	items, err := s.enrichItems(ctx, rows)
	if err != nil {
		return nil, err
	}
	summary, err := s.summarize([]UserList{*list}, "", 0)
	if err != nil {
		return nil, err
	}
	return &ListDetailResponse{
		List:  *summary,
		Items: items,
		Pagination: Pagination{
			Page: page, PageSize: ListItemsPageSize, Total: total,
			HasMore: int64(page*ListItemsPageSize) < total,
		},
	}, nil
}

func (s *Service) Update(userID, listID uint, req UpdateListRequest) (*ListSummary, error) {
	list, err := s.repo.updateList(listID, userID, func(tx *gorm.DB, list *UserList) (map[string]any, error) {
		updates := map[string]any{}
		if req.Name != nil {
			name, err := validName(*req.Name)
			if err != nil {
				return nil, err
			}
			updates["name"] = name
		}
		if req.Description != nil {
			desc, err := validDescription(*req.Description)
			if err != nil {
				return nil, err
			}
			updates["description"] = desc
		}
		if req.Visibility != nil {
			if err := validVisibility(*req.Visibility); err != nil {
				return nil, err
			}
			updates["visibility"] = *req.Visibility
		}
		if req.CoverItemID != nil {
			if *req.CoverItemID == 0 {
				updates["cover_item_id"] = nil
			} else {
				var n int64
				if err := tx.Model(&UserListItem{}).Where("id = ? AND list_id = ?", *req.CoverItemID, list.ID).Count(&n).Error; err != nil {
					return nil, err
				}
				if n == 0 {
					return nil, ErrItemNotFound
				}
				updates["cover_item_id"] = *req.CoverItemID
			}
		}
		if len(updates) == 0 {
			return nil, ErrInvalidInput
		}
		return updates, nil
	})
	if err != nil {
		return nil, err
	}
	return s.summarize([]UserList{*list}, "", 0)
}

func (s *Service) Delete(userID, listID uint) error { return s.repo.deleteList(listID, userID) }

func (s *Service) AddItem(ctx context.Context, userID, listID uint, req AddItemRequest) (*ListItemResponse, error) {
	if err := validMedia(req); err != nil {
		return nil, err
	}
	item, err := s.repo.addItem(listID, userID, req.MediaType, req.MediaID)
	if err != nil {
		return nil, err
	}
	items, err := s.enrichItems(ctx, []UserListItem{*item})
	if err != nil {
		return nil, err
	}
	return &items[0], nil
}

func (s *Service) RemoveItem(userID, listID, itemID uint) error {
	return s.repo.removeItem(listID, userID, itemID)
}

func (s *Service) Reorder(userID, listID uint, req ReorderRequest) error {
	return s.repo.reorder(listID, userID, req.ItemIDs)
}

func mapPolicyErr(err error) error {
	if errors.Is(err, privacy_policy.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

// summarize wraps summarizeAll for a single list.
func (s *Service) summarize(lists []UserList, mediaType string, mediaID int) (*ListSummary, error) {
	out, err := s.summarizeAll(lists, mediaType, mediaID)
	if err != nil || len(out) == 0 {
		return nil, err
	}
	return &out[0], nil
}

func (s *Service) summarizeAll(lists []UserList, mediaType string, mediaID int) ([]ListSummary, error) {
	data, err := s.repo.summaryData(lists, mediaType, mediaID)
	if err != nil {
		return nil, err
	}

	// Posters needed: the picked cover item, else the collage items. Fetched once per title.
	type ref struct {
		t  string
		id int
	}
	posters := map[ref]string{}
	for _, l := range lists {
		if c, ok := data.covers[l.ID]; ok {
			posters[ref{c.MediaType, c.MediaID}] = ""
		} else {
			for _, it := range data.collages[l.ID] {
				posters[ref{it.MediaType, it.MediaID}] = ""
			}
		}
	}
	refs := make([]ref, 0, len(posters))
	for r := range posters {
		refs = append(refs, r)
	}
	fetched := make([]string, len(refs))
	var g errgroup.Group
	g.SetLimit(tmdbConcurrency)
	for i, r := range refs {
		g.Go(func() error {
			fetched[i] = shared.FetchMediaSummary(r.id, shared.MediaType(r.t)).PosterURL
			return nil
		})
	}
	_ = g.Wait()
	for i, r := range refs {
		posters[r] = fetched[i]
	}

	out := make([]ListSummary, len(lists))
	for i, l := range lists {
		sum := ListSummary{
			ID: l.ID, UserID: l.UserID, Name: l.Name, Description: l.Description,
			Visibility: l.Visibility, ItemCount: data.counts[l.ID], CoverItemID: l.CoverItemID,
			CollagePosters: []string{}, CreatedAt: l.CreatedAt, UpdatedAt: l.UpdatedAt,
		}
		if c, ok := data.covers[l.ID]; ok {
			sum.CoverPoster = posters[ref{c.MediaType, c.MediaID}]
		} else {
			for _, it := range data.collages[l.ID] {
				if p := posters[ref{it.MediaType, it.MediaID}]; p != "" {
					sum.CollagePosters = append(sum.CollagePosters, p)
				}
			}
		}
		if mediaID > 0 && mediaType != "" {
			itemID, has := data.contains[l.ID]
			sum.ContainsMedia = &has
			if has {
				sum.ItemID = &itemID
			}
		}
		out[i] = sum
	}
	return out, nil
}

func (s *Service) enrichItems(ctx context.Context, rows []UserListItem) ([]ListItemResponse, error) {
	out := make([]ListItemResponse, len(rows))
	var g errgroup.Group
	g.SetLimit(tmdbConcurrency)
	for i := range rows {
		g.Go(func() error {
			out[i] = ListItemResponse{
				ID: rows[i].ID, Position: rows[i].Position, AddedAt: rows[i].CreatedAt,
				MediaType: shared.MediaType(rows[i].MediaType),
				Media:     shared.FetchMediaSummary(rows[i].MediaID, shared.MediaType(rows[i].MediaType)),
			}
			return nil
		})
	}
	_ = g.Wait()
	media := make([]*shared.MediaSummary, len(out))
	for i := range out {
		media[i] = &out[i].Media
	}
	if err := shared.AttachRatings(ctx, s.db, media); err != nil {
		return nil, err
	}
	return out, nil
}
