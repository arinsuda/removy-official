package manga_module

import (
	"html"
	"net/url"
	"regexp"
	"strings"
)

// AniList values used when mapping.
const (
	aniListManga       = "MANGA"
	aniListAnime       = "ANIME"
	aniListAdaptation  = "ADAPTATION"
	aniListFormatNovel = "NOVEL"
	linkTypeStreaming  = "STREAMING"
	linkTypeInfo       = "INFO"
	countryTW          = "TW"

	mediaTypeManga = "manga"
	// tagMinRank hides AniList tags that describe the title only weakly.
	tagMinRank = 60
	// aniListScoreScale is AniList's averageScore range (0-100).
	aniListScoreScale = 100
)

var (
	lineBreakTag = regexp.MustCompile(`(?i)<br\s*/?>`)
	anyTag       = regexp.MustCompile(`<[^>]*>`)
	thaiScript   = regexp.MustCompile(`\p{Thai}`)
)

// toItem maps one AniList title to the list DTO.
func toItem(m rawMedia) MangaItem {
	titles := Titles{
		Romaji:  m.Title.Romaji,
		English: m.Title.English,
		Native:  m.Title.Native,
		Thai:    thaiSynonym(m.Synonyms),
	}
	var cover string
	if m.CoverImage != nil {
		cover = httpsURL(m.CoverImage.Large)
	}
	return MangaItem{
		ID:           m.ID,
		MediaType:    mediaTypeManga,
		Titles:       titles,
		DisplayTitle: displayTitle(titles),
		CoverURL:     cover,
		BannerURL:    httpsURL(m.BannerImage),
		Format:       m.Format,
		Country:      m.CountryOfOrigin,
		Kind:         deriveKind(m.Format, m.CountryOfOrigin),
		IsAdult:      m.IsAdult,
		Status:       m.Status,
		StartYear:    startYear(m.StartDate),
		Chapters:     m.Chapters,
		Volumes:      m.Volumes,
		Genres:       nonNilStrings(m.Genres),
		Ratings:      ratingsFor(m.AverageScore),
	}
}

// toDetail maps the full AniList title, dropping spoiler tags and unsafe links.
func toDetail(m rawMedia) MangaDetail {
	var description string
	if m.Description != nil {
		description = plainText(*m.Description)
	}
	return MangaDetail{
		MangaItem:   toItem(m),
		Description: description,
		Tags:        readableTags(m.Tags),
		ReadLinks:   readLinks(m.ExternalLinks),
		Relations:   animeAdaptations(m.Relations),
	}
}

// toSimilar maps recommendations, dropping anything the viewer may not see.
func toSimilar(recs *rawRecommend, adult bool) []MangaItem {
	items := []MangaItem{}
	if recs == nil {
		return items
	}
	for _, n := range recs.Nodes {
		if m := n.MediaRecommendation; m != nil && isVisible(*m, adult) {
			items = append(items, toItem(*m))
		}
	}
	return items
}

// isVisible is a defence-in-depth check behind the query-level type filter and the isAdult filter.
// That filter is omitted when adult titles are included, so adult is what allows adult items through.
func isVisible(m rawMedia, adult bool) bool {
	return m.Type == aniListManga && (adult || !m.IsAdult)
}

// displayTitle picks the first non-empty title in the order Thai > English > Romaji > Native.
func displayTitle(t Titles) string {
	for _, s := range []string{t.Thai, t.English, t.Romaji, t.Native} {
		if s != "" {
			return s
		}
	}
	return ""
}

// thaiSynonym returns the first synonym written in Thai script.
func thaiSynonym(synonyms []string) string {
	for _, s := range synonyms {
		if thaiScript.MatchString(s) {
			return s
		}
	}
	return ""
}

// deriveKind classifies the title: novels by format, otherwise by country of origin.
func deriveKind(format, country string) string {
	switch {
	case format == aniListFormatNovel:
		return kindNovel
	case country == countryKR:
		return kindManhwa
	case country == countryCN || country == countryTW:
		return kindManhua
	default:
		return kindManga
	}
}

// plainText turns AniList's description into text: line breaks become newlines and every tag is removed,
// including tags that only appear after entity decoding.
func plainText(s string) string {
	s = lineBreakTag.ReplaceAllString(s, "\n")
	s = anyTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = anyTag.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// httpsURL returns the URL only when it is an absolute https link; anything else becomes "".
func httpsURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return ""
	}
	return u.String()
}

func startYear(d *rawDate) *int {
	if d == nil {
		return nil
	}
	return d.Year
}

func ratingsFor(score *int) Ratings {
	r := AniListRating{Scale: aniListScoreScale}
	if score != nil && *score > 0 {
		r.Average = *score
		r.Available = true
	}
	return Ratings{AniList: r}
}

func readableTags(tags []rawTag) []Tag {
	out := []Tag{}
	for _, t := range tags {
		if t.Rank < tagMinRank || t.IsMediaSpoiler || t.IsGeneralSpoiler {
			continue
		}
		out = append(out, Tag{Name: t.Name, Rank: t.Rank})
	}
	return out
}

// readLinks keeps official reading and information links that are https.
func readLinks(links []rawLink) []ReadLink {
	out := []ReadLink{}
	for _, l := range links {
		if l.Type != linkTypeStreaming && l.Type != linkTypeInfo {
			continue
		}
		safe := httpsURL(l.URL)
		if safe == "" {
			continue
		}
		out = append(out, ReadLink{Site: l.Site, URL: safe, Language: l.Language})
	}
	return out
}

// animeAdaptations lists anime made from this title (ADAPTATION edges to ANIME nodes).
func animeAdaptations(rel *rawEdges) []AnimeLink {
	out := []AnimeLink{}
	if rel == nil {
		return out
	}
	for _, e := range rel.Edges {
		if e.RelationType != aniListAdaptation || e.Node == nil || e.Node.Type != aniListAnime {
			continue
		}
		out = append(out, AnimeLink{ID: e.Node.ID, Title: englishFirst(e.Node.Title)})
	}
	return out
}

func englishFirst(t rawTitle) string {
	for _, s := range []string{t.English, t.Romaji, t.Native} {
		if s != "" {
			return s
		}
	}
	return ""
}

func nonNilStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
