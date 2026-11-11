package stremio

import (
	"testing"

	"github.com/braedonsaunders/ether/internal/core"
)

func TestAskableSeasonTakesSingleSpecials(t *testing.T) {
	for _, c := range []struct {
		req  core.SearchRequest
		want bool
	}{
		{core.SearchRequest{Type: core.TypeEpisode, Season: 0, Episode: 3}, true},
		{core.SearchRequest{Type: core.TypeEpisode, Season: 2, Episode: 3}, true},
		{core.SearchRequest{Type: core.TypeSeason, Season: 2}, true},
		{core.SearchRequest{Type: core.TypeSeason, Season: 0}, false},
		{core.SearchRequest{Type: core.TypeEpisode, Season: 0, Episode: 0}, false},
	} {
		if got := askableSeason(c.req); got != c.want {
			t.Errorf("askableSeason(%+v) = %v, want %v", c.req, got, c.want)
		}
	}
}
