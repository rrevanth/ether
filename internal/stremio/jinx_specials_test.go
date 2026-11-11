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

func TestSpecialInsideAPackTakesTheAddonsEpisode(t *testing.T) {
	req := core.SearchRequest{Type: core.TypeEpisode, Title: "Haikyu!!", Season: 0, Episode: 1, Anime: true}
	idx := 31
	streams := []stream{{
		Title:    "[Anime Time] Haikyuu!! (S01+02+03+04+Movies+OVAs+OADs) [BD] [Dual Audio][1080p]\n[Anime Time] Haikyuu!! OVA/[Anime Time] Haikyuu!! - OVA - Lev Appears.mkv\n👤 12 💾 290 MB ⚙️ NyaaSi",
		InfoHash: "0123456789abcdef0123456789abcdef01234567",
		FileIdx:  &idx,
	}}
	rels, _, _ := convertWith(streams, req, false, convertOptions{})
	if len(rels) != 1 {
		t.Fatalf("got %d releases", len(rels))
	}
	r := rels[0]
	if len(r.Info.Seasons) != 1 || r.Info.Seasons[0] != 0 || len(r.Info.Episodes) != 1 || r.Info.Episodes[0] != 1 || r.Info.FullSeason {
		t.Fatalf("identity = seasons %v episodes %v full %v, want S00E01", r.Info.Seasons, r.Info.Episodes, r.Info.FullSeason)
	}
	if r.FileName != "[Anime Time] Haikyuu!! - OVA - Lev Appears" {
		t.Fatalf("FileName = %q", r.FileName)
	}
}

func TestSpecialKeepsAFileThatNamesAnotherEpisode(t *testing.T) {
	req := core.SearchRequest{Type: core.TypeEpisode, Title: "Show", Season: 0, Episode: 1}
	streams := []stream{{
		Title:    "Show S01 1080p WEB-DL\nShow.S01E05.1080p.WEB-DL.mkv\n👤 3 💾 1 GB",
		InfoHash: "89abcdef0123456789abcdef0123456789abcdef",
	}}
	rels, _, _ := convertWith(streams, req, false, convertOptions{})
	if len(rels) != 1 || len(rels[0].Info.Episodes) != 1 || rels[0].Info.Episodes[0] != 5 {
		t.Fatalf("a file naming S01E05 was re-labelled: %+v", rels)
	}
}
