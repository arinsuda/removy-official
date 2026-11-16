package manga_module

import (
	"errors"
	"strconv"

	anilist "github.com/arinsuda/movie-hub/internal/anilist_module"
	"github.com/arinsuda/movie-hub/internal/shared"
	"github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
)

// Error messages are for logs and developers; clients branch on the code.
const (
	msgInvalidSearch = "query or filter parameter is missing or out of range"
	msgInvalidID     = "id must be a positive integer"
	msgNotFound      = "not found"
	msgUpstream      = "failed to load data from AniList"
	msgAgeRestricted = "adult titles need a signed-in account whose age is at least the minimum"
	msgAgeCheck      = "failed to check the viewer's age"
)

type Handler struct {
	svc  *MangaService
	gate AdultGate
}

func NewHandler(svc *MangaService, gate AdultGate) *Handler {
	return &Handler{svc: svc, gate: gate}
}

// New builds the handler shared by RegisterPublicRoutes; client is the only AniList caller.
func New(client *anilist.Client, gate AdultGate) *Handler {
	return NewHandler(NewMangaService(client), gate)
}

// adultAllowed asks the gate whether the request's viewer may see adult titles (guests never may).
func (h *Handler) adultAllowed(c fiber.Ctx) (bool, error) {
	var viewer uint
	if claims := middleware.GetClaims(c); claims != nil {
		viewer = claims.UserID
	}
	return h.gate.Eligible(c.Context(), viewer)
}

// List serves GET /manga (browse with filters).
func (h *Handler) List(c fiber.Ctx) error {
	f, ok := parseListFilter(c)
	if !ok {
		return invalidSearch(c)
	}
	eligible, err := h.adultAllowed(c)
	if err != nil {
		return ageCheckError(c)
	}
	f.Adult = f.IncludeAdult && eligible
	result, err := h.svc.List(c.Context(), f)
	if err != nil {
		return upstreamError(c, err)
	}
	return c.JSON(result)
}

// Search serves GET /manga/search?q=&page=.
func (h *Handler) Search(c fiber.Ctx) error {
	q, ok := searchQuery(c)
	if !ok {
		return invalidSearch(c)
	}
	page, ok := pageParam(c)
	if !ok {
		return invalidSearch(c)
	}
	includeAdult, ok := includeAdultParam(c)
	if !ok {
		return invalidSearch(c)
	}
	eligible, err := h.adultAllowed(c)
	if err != nil {
		return ageCheckError(c)
	}
	result, err := h.svc.Search(c.Context(), q, page, includeAdult && eligible)
	if err != nil {
		return upstreamError(c, err)
	}
	return c.JSON(result)
}

// Genres serves GET /manga/genres from the static allowlist, so it never calls AniList.
func (h *Handler) Genres(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"genres": Genres()})
}

// GetByID serves GET /manga/:id.
func (h *Handler) GetByID(c fiber.Ctx) error {
	id, ok := idParam(c)
	if !ok {
		return invalidID(c)
	}
	eligible, err := h.adultAllowed(c)
	if err != nil {
		return ageCheckError(c)
	}
	detail, err := h.svc.GetByID(c.Context(), id, eligible)
	if err != nil {
		return upstreamError(c, err)
	}
	return c.JSON(fiber.Map{"manga": detail})
}

// GetSimilar serves GET /manga/:id/similar?include_adult=. Adult recommendations need both the request
// flag and an eligible viewer, the same rule as search.
func (h *Handler) GetSimilar(c fiber.Ctx) error {
	id, ok := idParam(c)
	if !ok {
		return invalidID(c)
	}
	includeAdult, ok := includeAdultParam(c)
	if !ok {
		return invalidSearch(c)
	}
	eligible, err := h.adultAllowed(c)
	if err != nil {
		return ageCheckError(c)
	}
	items, err := h.svc.Similar(c.Context(), id, eligible, includeAdult && eligible)
	if err != nil {
		return upstreamError(c, err)
	}
	return c.JSON(fiber.Map{"results": items})
}

func idParam(c fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	return id, err == nil && id > 0
}

// upstreamError maps service failures: a missing title is 404, an adult title for an ineligible
// viewer is 403, everything else is 503.
func upstreamError(c fiber.Ctx, err error) error {
	if errors.Is(err, anilist.ErrNotFound) {
		return shared.WriteAPIError(c, fiber.StatusNotFound, shared.ErrorCodeNotFound, msgNotFound)
	}
	if errors.Is(err, ErrAgeRestricted) {
		return shared.WriteAPIError(c, fiber.StatusForbidden, shared.ErrorCodeAgeRestricted, msgAgeRestricted)
	}
	return shared.WriteAPIError(c, fiber.StatusServiceUnavailable, shared.ErrorCodeUpstreamUnavailable, msgUpstream)
}

func invalidSearch(c fiber.Ctx) error {
	return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidSearchParams, msgInvalidSearch)
}

func ageCheckError(c fiber.Ctx) error {
	return shared.WriteAPIError(c, fiber.StatusInternalServerError, shared.ErrorCodeInternal, msgAgeCheck)
}

func invalidID(c fiber.Ctx) error {
	return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidID, msgInvalidID)
}
