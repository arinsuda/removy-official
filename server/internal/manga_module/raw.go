package manga_module

// Raw AniList payloads. Only the fields the queries ask for are declared.

type rawMedia struct {
	ID              int           `json:"id"`
	Type            string        `json:"type"`
	IsAdult         bool          `json:"isAdult"`
	Title           rawTitle      `json:"title"`
	Synonyms        []string      `json:"synonyms"`
	CoverImage      *rawCover     `json:"coverImage"`
	BannerImage     string        `json:"bannerImage"`
	Format          string        `json:"format"`
	CountryOfOrigin string        `json:"countryOfOrigin"`
	Status          string        `json:"status"`
	StartDate       *rawDate      `json:"startDate"`
	Chapters        *int          `json:"chapters"`
	Volumes         *int          `json:"volumes"`
	Genres          []string      `json:"genres"`
	AverageScore    *int          `json:"averageScore"`
	Description     *string       `json:"description"`
	Tags            []rawTag      `json:"tags"`
	ExternalLinks   []rawLink     `json:"externalLinks"`
	Relations       *rawEdges     `json:"relations"`
	Recommendations *rawRecommend `json:"recommendations"`
}

type rawTitle struct {
	Romaji  string `json:"romaji"`
	English string `json:"english"`
	Native  string `json:"native"`
}

type rawCover struct {
	Large string `json:"large"`
}

type rawDate struct {
	Year *int `json:"year"`
}

type rawTag struct {
	Name             string `json:"name"`
	Rank             int    `json:"rank"`
	IsMediaSpoiler   bool   `json:"isMediaSpoiler"`
	IsGeneralSpoiler bool   `json:"isGeneralSpoiler"`
}

type rawLink struct {
	Site     string `json:"site"`
	URL      string `json:"url"`
	Type     string `json:"type"`
	Language string `json:"language"`
}

type rawEdges struct {
	Edges []rawEdge `json:"edges"`
}

type rawEdge struct {
	RelationType string    `json:"relationType"`
	Node         *rawMedia `json:"node"`
}

type rawRecommend struct {
	Nodes []rawRecommendation `json:"nodes"`
}

type rawRecommendation struct {
	MediaRecommendation *rawMedia `json:"mediaRecommendation"`
}

// rawPageInfo is the Page.pageInfo selection.
type rawPageInfo struct {
	Total       int `json:"total"`
	CurrentPage int `json:"currentPage"`
	LastPage    int `json:"lastPage"`
}
