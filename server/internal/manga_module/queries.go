package manga_module

// Every query hard-codes type: MANGA in the query text. isAdult is a variable on browse and search:
// false for everyone who may not see adult titles, and omitted (no filter) for an eligible viewer who
// asked for them, because isAdult:true would return adult titles only. Single-title and recommendation
// lookups filter nothing by isAdult and gate the returned isAdult field after the cache. Unset filters
// are omitted.

// mangaFields is the shared selection for one title.
const mangaFields = `
fragment MangaFields on Media {
  id
  type
  isAdult
  title { romaji english native }
  synonyms
  coverImage { large }
  bannerImage
  format
  countryOfOrigin
  status
  startDate { year }
  chapters
  volumes
  genres
  averageScore
}`

// listQuery serves browse (filters) and search (search variable + SEARCH_MATCH sort).
const listQuery = mangaFields + `
query ($page: Int, $perPage: Int, $search: String, $isAdult: Boolean, $country: CountryCode, $format_in: [MediaFormat], $genre: String, $status: MediaStatus, $sort: [MediaSort]) {
  Page(page: $page, perPage: $perPage) {
    pageInfo { total currentPage lastPage }
    media(type: MANGA, isAdult: $isAdult, search: $search, countryOfOrigin: $country, format_in: $format_in, genre: $genre, status: $status, sort: $sort) {
      ...MangaFields
    }
  }
}`

// detailQuery loads one title with its long-form fields.
const detailQuery = mangaFields + `
query ($id: Int) {
  Media(id: $id, type: MANGA) {
    ...MangaFields
    description(asHtml: false)
    tags { name rank isMediaSpoiler isGeneralSpoiler }
    externalLinks { site url type language }
    relations {
      edges {
        relationType(version: 2)
        node { id type title { romaji english native } }
      }
    }
  }
}`

// similarQuery returns AniList recommendations; the caller still drops anything not MANGA or adult.
const similarQuery = mangaFields + `
query ($id: Int, $perPage: Int) {
  Media(id: $id, type: MANGA) {
    recommendations(sort: RATING_DESC, perPage: $perPage) {
      nodes { mediaRecommendation { ...MangaFields } }
    }
  }
}`

// sortSearchMatch ranks search hits by how well they match the query text.
const sortSearchMatch = "SEARCH_MATCH"

// listVariables builds the browse variables; unset filters are left out so AniList applies no filter.
func listVariables(f ListFilter) map[string]any {
	vars := withAdultFilter(map[string]any{
		"page":      f.Page,
		"perPage":   mangaPageSize,
		"sort":      []string{f.Sort},
		"format_in": f.Formats,
	}, f.Adult)
	if f.Country != "" {
		vars["country"] = f.Country
	}
	if f.Genre != "" {
		vars["genre"] = f.Genre
	}
	if f.Status != "" {
		vars["status"] = f.Status
	}
	return vars
}

// searchVariables builds the search variables (no format, country, genre or status filters).
func searchVariables(q string, page int, adult bool) map[string]any {
	return withAdultFilter(map[string]any{
		"search":  q,
		"page":    page,
		"perPage": mangaPageSize,
		"sort":    []string{sortSearchMatch},
	}, adult)
}

// withAdultFilter sends isAdult:false unless adult titles are included; when they are, the variable is
// left out so AniList returns both adult and non-adult titles.
func withAdultFilter(vars map[string]any, adult bool) map[string]any {
	if !adult {
		vars["isAdult"] = false
	}
	return vars
}

// idVariables builds variables for a single title.
func idVariables(id int) map[string]any {
	return map[string]any{"id": id}
}

// similarVariables builds variables for the recommendations lookup.
func similarVariables(id int) map[string]any {
	return map[string]any{"id": id, "perPage": similarPerPage}
}
