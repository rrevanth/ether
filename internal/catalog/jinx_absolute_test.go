package catalog

import (
	"context"
	"testing"
	"time"

	"github.com/braedonsaunders/ether/internal/core"
)

// One Piece's TMDB seasons number episodes from the series start: S22
// (Egghead, 67 episodes after 1088) is episodes 1089-1155. Its absolute
// number is the episode number, not 1088 more. A number within the season's
// own count (S02E05 of a 16-episode season) is relative and adds the earlier
// seasons as before.
func TestHydrateSeasonsNumberedFromTheSeriesStart(t *testing.T) {
	ctx := context.Background()
	c := newChunkTestCatalog(t)
	aired := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
	op := core.Title{Key: core.ShowKey(37854), Type: core.TypeShow, Title: "One Piece", Released: aired}
	season := func(n, count int) core.Title {
		return core.Title{Key: core.SeasonKey(37854, n), Type: core.TypeSeason, ShowKey: op.Key, Season: n, ChildCount: count, Title: "Season", Released: aired}
	}
	if err := c.upsertMany(ctx, []core.Title{op, season(1, 61), season(2, 16), season(3, 1011), season(4, 67), season(5, 25)}, upsertOpts{browsable: true}); err != nil {
		t.Fatal(err)
	}
	ep := func(s, e int) core.Title {
		return core.Title{Key: core.EpisodeKey(37854, s, e), Type: core.TypeEpisode, ShowKey: op.Key, Season: s, Episode: e}
	}
	items := []core.Title{ep(4, 1100), ep(5, 1180), ep(2, 62), ep(2, 5)}
	if err := c.hydrateSlice(ctx, items); err != nil {
		t.Fatal(err)
	}
	for i, want := range []int{1100, 1180, 62, 61 + 5} {
		if got := items[i].AbsoluteEpisode; got != want {
			t.Errorf("S%dE%d: absolute %d, want %d", items[i].Season, items[i].Episode, got, want)
		}
	}
}
