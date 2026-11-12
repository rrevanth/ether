package catalog

import (
	"context"
	"testing"
	"time"

	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/tmdb"
)

func day(s string) time.Time { t, _ := time.Parse("2006-01-02", s); return t }

// One Piece S22 (Egghead) has no TMDB air date; its episodes do.
func TestUndatedSeasonTakesItsFirstEpisodeDate(t *testing.T) {
	s := datedFromEpisodes(core.Title{Type: core.TypeSeason, Season: 22}, []core.Title{{Released: day("2024-01-14")}, {Released: day("2024-01-07")}, {}})
	if !s.Released.Equal(day("2024-01-07")) || s.Year != 2024 {
		t.Fatalf("season dated %v (%d)", s.Released, s.Year)
	}
	dated := datedFromEpisodes(core.Title{Type: core.TypeSeason, Season: 3, Released: day("2001-01-01")}, []core.Title{{Released: day("2000-01-01")}})
	if !dated.Released.Equal(day("2001-01-01")) {
		t.Fatal("a TMDB season date was replaced")
	}
	if sp := datedFromEpisodes(core.Title{Type: core.TypeSeason, Season: 0}, []core.Title{{Released: day("2000-01-01")}}); !sp.Released.IsZero() {
		t.Fatal("specials are upstream's to date")
	}
}

// Undated seasons up to the last aired one list (provisionally dated by the
// previous season); a future undated season stays unreleased.
func TestUndatedSummariesUpToTheLastAiredSeasonList(t *testing.T) {
	c := &Catalog{}
	show := core.Title{Key: 2000000000037854, Released: day("1999-10-20")}
	seasons := []core.Title{
		{Season: 1, ChildCount: 61, Released: day("1999-10-20")},
		{Season: 2, ChildCount: 16},
		{Season: 21, ChildCount: 197, Released: day("2019-07-07")},
		{Season: 22, ChildCount: 67},
		{Season: 24, ChildCount: 10},
		{Season: 0, ChildCount: 39},
	}
	c.datedSummaries(context.Background(), show, &tmdb.TVDetails{LastEpisodeToAir: &tmdb.Episode{SeasonNumber: 23}}, seasons)
	want := map[int]time.Time{1: day("1999-10-20"), 2: day("1999-10-20"), 21: day("2019-07-07"), 22: day("2019-07-07"), 24: {}, 0: {}}
	for _, s := range seasons {
		if !s.Released.Equal(want[s.Season]) {
			t.Errorf("season %d dated %v, want %v", s.Season, s.Released, want[s.Season])
		}
	}
}
