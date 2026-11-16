package manga_module

// PaginatedResult mirrors the tmdb list envelope (results, page, total_pages, total_results).
type PaginatedResult[T any] struct {
	Page         int `json:"page"`
	Results      []T `json:"results"`
	TotalPages   int `json:"total_pages"`
	TotalResults int `json:"total_results"`
}

// Titles keeps every known title; Thai comes from AniList synonyms written in Thai script.
type Titles struct {
	Romaji  string `json:"romaji"`
	English string `json:"english"`
	Native  string `json:"native"`
	Thai    string `json:"thai"`
}

// AniListRating is AniList's average score on a 0-100 scale; Available is false without a score.
type AniListRating struct {
	Average   int  `json:"average"`
	Available bool `json:"available"`
	Scale     int  `json:"scale"`
}

// Ratings groups score sources. Only AniList is provided until POPCORN reviews accept manga.
type Ratings struct {
	AniList AniListRating `json:"anilist"`
}

// MangaItem is one title in browse, search and similar lists.
type MangaItem struct {
	ID           int      `json:"id"`
	MediaType    string   `json:"media_type"`
	Titles       Titles   `json:"titles"`
	DisplayTitle string   `json:"display_title"`
	CoverURL     string   `json:"cover_url"`
	BannerURL    string   `json:"banner_url"`
	Format       string   `json:"format"`
	Country      string   `json:"country"`
	Kind         string   `json:"kind"`
	IsAdult      bool     `json:"is_adult"`
	Status       string   `json:"status"`
	StartYear    *int     `json:"start_year"`
	Chapters     *int     `json:"chapters"`
	Volumes      *int     `json:"volumes"`
	Genres       []string `json:"genres"`
	Ratings      Ratings  `json:"ratings"`
}

// MangaDetail extends the list item with the long-form fields.
type MangaDetail struct {
	MangaItem
	Description string      `json:"description"`
	Tags        []Tag       `json:"tags"`
	ReadLinks   []ReadLink  `json:"read_links"`
	Relations   []AnimeLink `json:"relations"`
}

// Tag is a non-spoiler AniList tag at or above the rank threshold.
type Tag struct {
	Name string `json:"name"`
	Rank int    `json:"rank"`
}

// ReadLink points at an official or licensed place to read the title (https only).
type ReadLink struct {
	Site     string `json:"site"`
	URL      string `json:"url"`
	Language string `json:"language"`
}

// AnimeLink is an anime adaptation of the title.
type AnimeLink struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}
