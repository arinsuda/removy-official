package tmdbmodule

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapWatchProviders(t *testing.T) {
	var raw rawWatchProviders
	body := `{"results":{"TH":{"link":"https://www.themoviedb.org/movie/1/watch?locale=TH",
	"flatrate":[{"provider_id":2,"provider_name":"B","logo_path":"/b.jpg","display_priority":5},
	{"provider_id":1,"provider_name":"A","logo_path":"/a.jpg","display_priority":1}],
	"rent":[{"provider_id":3,"provider_name":"C","logo_path":"/c.jpg","display_priority":2}]}}}`
	assert.NoError(t, json.Unmarshal([]byte(body), &raw))

	got := mapWatchProviders(raw, "TH")
	assert.Equal(t, "TH", got.Region)
	assert.Equal(t, []int{1, 2}, []int{got.Flatrate[0].ProviderID, got.Flatrate[1].ProviderID})
	assert.Len(t, got.Rent, 1)
	assert.NotNil(t, got.Free)
	assert.NotNil(t, got.Ads)
	assert.NotNil(t, got.Buy)
}

func TestMapWatchProvidersMissingRegion(t *testing.T) {
	got := mapWatchProviders(rawWatchProviders{}, "JP")
	assert.Equal(t, "JP", got.Region)
	assert.Equal(t, "", got.Link)
	b, _ := json.Marshal(got)
	assert.Contains(t, string(b), `"flatrate":[]`)
	assert.Contains(t, string(b), `"buy":[]`)
}
