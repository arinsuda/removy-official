package movie_module

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	tmdb "github.com/arinsuda/movie-hub/internal/tmdb_module"
	"github.com/gofiber/fiber/v3"
)

const (
	providersTTL      = 6 * time.Hour
	providersMaxItems = 2000
	defaultRegion     = "TH"
)

var regionRe = regexp.MustCompile(`^[A-Z]{2}$`)

type providersEntry struct {
	value   *tmdb.WatchProviders
	expires time.Time
}

// providersCache is a small bounded TTL map keyed type:id:region.
type providersCache struct {
	mu    sync.Mutex
	items map[string]providersEntry
	now   func() time.Time
}

func newProvidersCache() *providersCache {
	return &providersCache{items: map[string]providersEntry{}, now: time.Now}
}

func (p *providersCache) get(key string) (*tmdb.WatchProviders, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	e, ok := p.items[key]
	if !ok || !p.now().Before(e.expires) {
		delete(p.items, key)
		return nil, false
	}
	return e.value, true
}

func (p *providersCache) set(key string, v *tmdb.WatchProviders) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.items) >= providersMaxItems {
		now := p.now()
		for k, e := range p.items {
			if !now.Before(e.expires) {
				delete(p.items, k)
			}
		}
		for k := range p.items { // still full: drop an arbitrary entry
			if len(p.items) < providersMaxItems {
				break
			}
			delete(p.items, k)
		}
	}
	p.items[key] = providersEntry{value: v, expires: p.now().Add(providersTTL)}
}

// GetWatchProviders returns cached providers for kind ("movie" or "tv").
func (s *MovieService) GetWatchProviders(kind string, id int, region string) (*tmdb.WatchProviders, error) {
	key := fmt.Sprintf("%s:%d:%s", kind, id, region)
	if v, ok := s.providers.get(key); ok {
		return v, nil
	}
	res, err := s.fetchProviders(kind, id, region)
	if err != nil {
		return nil, err
	}
	s.providers.set(key, res)
	return res, nil
}

func (h *Handler) watchProviders(c fiber.Ctx, kind string) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}
	region := c.Query("region", defaultRegion)
	if !regionRe.MatchString(region) {
		return c.Status(400).JSON(fiber.Map{"error": "region ไม่ถูกต้อง"})
	}
	res, err := h.svc.GetWatchProviders(kind, id, region)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลแหล่งรับชมไม่สำเร็จ"})
	}
	return c.JSON(fiber.Map{"providers": res})
}

func (h *Handler) GetMovieWatchProviders(c fiber.Ctx) error  { return h.watchProviders(c, "movie") }
func (h *Handler) GetSeriesWatchProviders(c fiber.Ctx) error { return h.watchProviders(c, "tv") }
