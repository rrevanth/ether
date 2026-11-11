package resolver

import (
	"context"
	"testing"

	"github.com/braedonsaunders/ether/internal/core"
)

// The 2026 animated series' TMDB alternative title is plain "RAMAYAN"; a
// yearless "Ramayan.S01E01-44" belongs to the 1987 Ramayan the catalogue
// also holds. Its own full-title release still matches.
func TestAlternativeTitleNamingAnotherShowIsRejected(t *testing.T) {
	asked := map[string]int{}
	s := New(&fakeIndexer{}, fakeHealth{}, nil, testCfg(), Options{TitleNamesake: func(_ context.Context, _ core.RatingKey, title string) (bool, error) {
		asked[title]++
		return title == "Ramayan", nil
	}})
	title := &core.Title{Type: core.TypeEpisode, Title: "The Echo Of Ayodhya", ShowKey: 2000000000339052, ShowTitle: "RAMAYAN AN ATHARV SRIVASTAVA SERIES", ShowYear: 2026,
		ShowAlternativeTitles: []string{"RAMAYAN"}, Season: 1, Episode: 1}
	sreq := core.SearchRequest{Type: core.TypeEpisode, Title: title.ShowTitle, Year: 2026, Season: 1, Episode: 1, EpisodeTitle: title.Title, AlternativeTitles: []string{"RAMAYAN"}}
	namesake := withParsed(rel("a", "Ramayan.S01E01-44.1080p.DSNP+.WEB-DL.AAC.2.0.x264-Telly", 1400))
	own := withParsed(rel("b", "RAMAYAN AN ATHARV SRIVASTAVA SERIES S01E01", 900))
	withYear := withParsed(rel("c", "Ramayan 2026 S01E01 1080p WEB-DL", 900))
	s.prepareNamesakes(context.Background(), title, sreq, []core.Release{namesake, own, withYear})
	if got := s.rejectNamesake(namesake, title, sreq); got != "title names another show" {
		t.Fatalf("1987 Ramayan pack: %q", got)
	}
	if got := s.rejectNamesake(own, title, sreq); got != "" {
		t.Fatalf("own full title rejected: %q", got)
	}
	if got := s.rejectNamesake(withYear, title, sreq); got != "" {
		t.Fatalf("release naming the show's year rejected: %q", got)
	}
	if asked["Ramayan"] != 1 || len(asked) != 1 {
		t.Fatalf("catalogue asked %v, want once for Ramayan only", asked)
	}
}

// An alternative title that names no other show (a romanized original)
// keeps matching.
func TestAlternativeTitleWithoutNamesakeStillMatches(t *testing.T) {
	s := New(&fakeIndexer{}, fakeHealth{}, nil, testCfg(), Options{TitleNamesake: func(context.Context, core.RatingKey, string) (bool, error) { return false, nil }})
	title := &core.Title{Type: core.TypeEpisode, ShowKey: 2000000000209867, ShowTitle: "Frieren: Beyond Journey's End", ShowAlternativeTitles: []string{"Sousou no Frieren"}, Season: 1, Episode: 3}
	sreq := core.SearchRequest{Type: core.TypeEpisode, Title: title.ShowTitle, Season: 1, Episode: 3, AlternativeTitles: title.ShowAlternativeTitles}
	r := withParsed(rel("f", "[SubsPlease] Sousou no Frieren - 03 (1080p) [ABCD1234].mkv", 1400))
	s.prepareNamesakes(context.Background(), title, sreq, []core.Release{r})
	if got := s.rejectNamesake(r, title, sreq); got != "" {
		t.Fatalf("romanized title rejected: %q", got)
	}
}
