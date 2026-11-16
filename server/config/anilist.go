package config

import "time"

// AniList supplies manga, manhwa, manhua and light novel metadata (read-only GraphQL, no key).
//
// Terms of use: AniList is free to use while revenue stays under AniListFreeRevenueCapUSD per
// month. Above that a commercial licence is required; contact contact@anilist.co before passing it.
const (
	// AniListFreeRevenueCapUSD is the monthly revenue (USD) under which AniList use is free.
	AniListFreeRevenueCapUSD = 150

	// AniListGraphQLURL is the public GraphQL endpoint.
	AniListGraphQLURL = "https://graphql.anilist.co"
	// AniListRequestTimeout bounds one upstream HTTP call.
	AniListRequestTimeout = 8 * time.Second
	// AniListMaxResponseBytes caps how much of one upstream response is read.
	AniListMaxResponseBytes = 2 << 20

	// AniListDetailCacheTTL caches detail, similar and search-by-id responses.
	AniListDetailCacheTTL = 6 * time.Hour
	// AniListListCacheTTL caches browse and search list responses.
	AniListListCacheTTL = 30 * time.Minute
	// AniListNotFoundTTL caches an AniList "no such title" answer briefly, so repeated bad ids do not
	// spend the upstream budget again and again.
	AniListNotFoundTTL = 10 * time.Minute
	// AniListCacheMaxEntries bounds the in-process response cache.
	AniListCacheMaxEntries = 1000

	// MangaRateMax and MangaRateWindow limit every /api/manga request per signed-in user or, for guests,
	// per client IP. Signed-in users count too: each uncached call spends the shared AniList budget.
	MangaRateMax    = 30
	MangaRateWindow = time.Minute

	// AniListRequestsPerMinute is the local token-bucket budget (burst equals one minute).
	AniListRequestsPerMinute = 60
	// AniListRetryAfterDefault is the cooldown after a 429 that carries no usable Retry-After.
	AniListRetryAfterDefault = time.Minute
	// AniListRetryAfterMax caps a cooldown so one bad header cannot silence the feature for hours.
	AniListRetryAfterMax = 10 * time.Minute
)
