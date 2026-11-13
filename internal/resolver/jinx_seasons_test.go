package resolver

import (
	"slices"
	"testing"

	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/releaseparse"
)

func TestShowNamesKeepRomajiAndDropPartTitles(t *testing.T) {
	sreq := core.SearchRequest{Title: "Attack on Titan", OriginalTitle: "進撃の巨人", AlternativeTitles: []string{"Shingeki no Kyojin",
		"Shingeki no Kyojin Lost Girls", "Attack on Titan: Lost Girls", "AOT", "Attack on Titan: The Final Season"}}
	got := showNames(sreq)
	for _, want := range []string{"Attack on Titan", "進撃の巨人", "Shingeki no Kyojin", "AOT"} {
		if !slices.Contains(got, want) {
			t.Errorf("showNames lacks %q: %v", want, got)
		}
	}
	for _, part := range []string{"Shingeki no Kyojin Lost Girls", "Attack on Titan: Lost Girls", "Attack on Titan: The Final Season"} {
		if slices.Contains(got, part) {
			t.Errorf("showNames kept the part title %q", part)
		}
	}
	// With the real catalogue names, the Kawaiika OAD file names only a count.
	path := "[Kawaiika-Raws] Shingeki no Kyojin OAD [DVDRip 1024x576 HEVC FLAC]/[Kawaiika-Raws] Shingeki no Kyojin OAD 02 [DVDRip 1024x576 HEVC FLAC].mkv"
	if !releaseparse.SpecialCountOnly(path, got) {
		t.Error("count-only OAD file not recognised with the romaji show name")
	}
	lost := "Attack on Titan Season 0/Lost Girls/Shingeki no Kyojin - Lost Girls OVA - 01 (s00e06).mkv"
	if !releaseparse.SpecialNamesAnother(lost, "Chibi Theater: Fly, Cadets, Fly!: Day 11", []string{"Lost Girls (1)"}, got) {
		t.Error("Lost Girls file for a Chibi Theater special kept")
	}
}

func TestReleaseNamingAnotherSeasonIsRejected(t *testing.T) {
	seasons := map[int]string{0: "Specials", 1: "Bleach", 2: "Thousand-Year Blood War"}
	s1 := core.SearchRequest{Type: core.TypeEpisode, Title: "Bleach", Season: 1, Episode: 1, SeasonTitles: seasons}
	tybw := withParsed(rel("a", "Bleach - Thousand-Year Blood War - S01E01 - 2160p WEB H.264 -NanDesuKa (B-Global)", 1500))
	if got := rejectOtherSeasonOrSpecial(tybw, s1); got == "" {
		t.Fatalf("TYBW release kept for Bleach S01E01 (title %q)", tybw.Info.Title)
	}
	plain := withParsed(rel("b", "[Anime Time] Bleach - 001 - The Day I Became a Shinigami", 300))
	if got := rejectOtherSeasonOrSpecial(plain, s1); got != "" {
		t.Fatalf("Bleach 001 rejected: %q", got)
	}
	s2 := s1
	s2.Season = 2
	if got := rejectOtherSeasonOrSpecial(tybw, s2); got != "" {
		t.Fatalf("TYBW release rejected for its own season: %q", got)
	}
	ds := core.SearchRequest{Type: core.TypeEpisode, Title: "Demon Slayer: Kimetsu no Yaiba", Season: 2, Episode: 1,
		SeasonTitles: map[int]string{1: "Unwavering Resolve Arc", 2: "Mugen Train Arc", 3: "Entertainment District Arc"}}
	if got := rejectOtherSeasonOrSpecial(withParsed(rel("c", "Demon Slayer Kimetsu no Yaiba S02E01 Mugen Train Arc 1080p", 900)), ds); got != "" {
		t.Fatalf("own arc rejected: %q", got)
	}
}
