package tmdbmodule

import (
	"fmt"
	"sort"
)

// WatchProvider is one streaming/rental provider entry.
type WatchProvider struct {
	ProviderID      int    `json:"provider_id"`
	ProviderName    string `json:"provider_name"`
	LogoPath        string `json:"logo_path"`
	DisplayPriority int    `json:"display_priority"`
}

// WatchProviders is the per-region result; slices are never nil.
type WatchProviders struct {
	Region   string          `json:"region"`
	Link     string          `json:"link"`
	Flatrate []WatchProvider `json:"flatrate"`
	Free     []WatchProvider `json:"free"`
	Ads      []WatchProvider `json:"ads"`
	Rent     []WatchProvider `json:"rent"`
	Buy      []WatchProvider `json:"buy"`
}

type rawWatchRegion struct {
	Link     string          `json:"link"`
	Flatrate []WatchProvider `json:"flatrate"`
	Free     []WatchProvider `json:"free"`
	Ads      []WatchProvider `json:"ads"`
	Rent     []WatchProvider `json:"rent"`
	Buy      []WatchProvider `json:"buy"`
}

type rawWatchProviders struct {
	Results map[string]rawWatchRegion `json:"results"`
}

func sortedProviders(in []WatchProvider) []WatchProvider {
	out := make([]WatchProvider, len(in))
	copy(out, in)
	sort.SliceStable(out, func(i, j int) bool { return out[i].DisplayPriority < out[j].DisplayPriority })
	return out
}

// mapWatchProviders picks one region; a missing region yields empty lists.
func mapWatchProviders(raw rawWatchProviders, region string) *WatchProviders {
	r := raw.Results[region]
	return &WatchProviders{
		Region:   region,
		Link:     r.Link,
		Flatrate: sortedProviders(r.Flatrate),
		Free:     sortedProviders(r.Free),
		Ads:      sortedProviders(r.Ads),
		Rent:     sortedProviders(r.Rent),
		Buy:      sortedProviders(r.Buy),
	}
}

// GetWatchProviders fetches /{movie|tv}/{id}/watch/providers for one region.
func GetWatchProviders(kind string, tmdbID int, region string) (*WatchProviders, error) {
	var raw rawWatchProviders
	if err := get(fmt.Sprintf("/%s/%d/watch/providers", kind, tmdbID), nil, &raw); err != nil {
		return nil, err
	}
	return mapWatchProviders(raw, region), nil
}
