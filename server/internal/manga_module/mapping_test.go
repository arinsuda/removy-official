package manga_module

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisplayTitleFallbackOrder(t *testing.T) {
	cases := []struct {
		name string
		in   Titles
		want string
	}{
		{"thai wins", Titles{Thai: "ไทย", English: "En", Romaji: "Ro", Native: "Na"}, "ไทย"},
		{"english next", Titles{English: "En", Romaji: "Ro", Native: "Na"}, "En"},
		{"romaji next", Titles{Romaji: "Ro", Native: "Na"}, "Ro"},
		{"native last", Titles{Native: "Na"}, "Na"},
		{"nothing", Titles{}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, displayTitle(tc.in))
		})
	}
}

func TestThaiSynonymMustContainThaiScript(t *testing.T) {
	assert.Equal(t, "มังงะ", thaiSynonym([]string{"Manga", "มังงะ"}))
	assert.Equal(t, "", thaiSynonym([]string{"Manga", "Mangá"}))
}

func TestDeriveKind(t *testing.T) {
	cases := []struct{ format, country, want string }{
		{"NOVEL", "JP", kindNovel},
		{"NOVEL", "KR", kindNovel},
		{"MANGA", "KR", kindManhwa},
		{"MANGA", "CN", kindManhua},
		{"ONE_SHOT", "TW", kindManhua},
		{"MANGA", "JP", kindManga},
		{"ONE_SHOT", "JP", kindManga},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, deriveKind(tc.format, tc.country), tc.format+"/"+tc.country)
	}
}

func TestPlainTextStripsEveryTag(t *testing.T) {
	assert.Equal(t, "Hello world & x", plainText("<br><br>Hello <i>world</i> &amp; <b>x</b>"))
	assert.Equal(t, "alert(1)", plainText("&lt;script&gt;alert(1)&lt;/script&gt;"))
	assert.NotContains(t, plainText("&lt;img src=x&gt;"), "<")
}

func TestHTTPSURLOnly(t *testing.T) {
	assert.Equal(t, "https://a.example/x.png", httpsURL("https://a.example/x.png"))
	assert.Equal(t, "", httpsURL("http://a.example/x.png"))
	assert.Equal(t, "", httpsURL("javascript:alert(1)"))
	assert.Equal(t, "", httpsURL(""))
}

func TestToItemCarriesAdultFlagAndRatings(t *testing.T) {
	var m rawMedia
	require.NoError(t, json.Unmarshal([]byte(mediaFixture(adultTitleID, true)), &m))
	item := toItem(m)
	assert.True(t, item.IsAdult)
	assert.Equal(t, mediaTypeManga, item.MediaType)
	assert.Equal(t, aniListScoreScale, item.Ratings.AniList.Scale)
	assert.True(t, item.Ratings.AniList.Available)
	assert.Equal(t, 72, item.Ratings.AniList.Average)
}

func TestSimilarDropsNonManga(t *testing.T) {
	var m rawMedia
	require.NoError(t, json.Unmarshal([]byte(mediaFixture(1, false)), &m))
	recs := &rawRecommend{Nodes: []rawRecommendation{
		{MediaRecommendation: &m},
	}}
	assert.Len(t, toSimilar(recs, false), 1)
	recs.Nodes[0].MediaRecommendation.Type = "ANIME"
	assert.Empty(t, toSimilar(recs, true))
}

func TestNonAdultScoreAbsentIsUnavailable(t *testing.T) {
	r := ratingsFor(nil)
	assert.False(t, r.AniList.Available)
	assert.Equal(t, 0, r.AniList.Average)
}
