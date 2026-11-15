package shared

import "context"

// MediaAward is one award a title won, as shown on its movie or TV detail page.
type MediaAward struct {
	Kind   string `json:"kind"`
	Period string `json:"period"`
}

// AwardReader looks up the awards of titles. event_module.AwardReader implements it.
type AwardReader interface {
	AwardsFor(ctx context.Context, ids []MediaIdentity) (map[MediaKey][]MediaAward, error)
}
