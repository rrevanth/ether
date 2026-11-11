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

// Cinemeta's (IMDb's) season 0 for Haikyu!! and Kuroko's Basketball, as listed on 2026-11-11.
var haikyuIMDb = []specialVideo{{1, "The Arrival of Lev"}, {2, "Haikyu!! the Movie 1: The End and the Beginning"}, {3, "Vs. Failing Grades"},
	{4, "Haikyu!! the Movie 2: The Winner and the Loser"}, {5, "Special Feature! Betting on the Spring High Volleyball"}, {6, "Haikyu!! the Movie 3: Talent and Sense"},
	{7, "Haikyu!! the Movie 4: Battle of Concepts"}, {8, "Land vs. Air"}, {9, "The Path of the Ball"}, {11, "Haikyu!! Quest Picture Drama 01"},
	{12, "Haikyu!! Quest Picture Drama 02"}, {13, "Haikyu!! Quest Picture Drama 03"}}

var kurokoIMDb = []specialVideo{{1, "NG-Shuu 1"}, {2, "NG-Shuu 2"}, {7, "Let's Chat OVA"}, {12, "Idiots Can't Win!"}, {13, "NG-Shuu 2 1"},
	{23, "Would You Like To Chat? OVA"}, {24, "NG-Shuu 3 1"}, {37, "Kuroko's Basketball: Last Game"}, {38, "NG-Shuu Last Game"},
	{39, "Shall We Chat? OVA"}, {40, "The Best Present (Episode 75.5)"}}

func TestMatchSpecialRenumbersTMDBToIMDb(t *testing.T) {
	for _, c := range []struct {
		videos  []specialVideo
		title   string
		episode int
		want    int // 0: not mapped
	}{
		{haikyuIMDb, "Lev Appears!", 1, 1},
		{haikyuIMDb, "VS Failing Marks", 2, 3},
		{haikyuIMDb, "Special Feature! The Spring Tournament of Their Youth", 3, 5},
		{haikyuIMDb, "Land vs. Air", 4, 8},
		{haikyuIMDb, "The Path of the Ball", 5, 9},
		{haikyuIMDb, "Haikyu Quest Picture Drama - Episode 1: Before the Boss Battle", 6, 11},
		{haikyuIMDb, "Haikyu Quest Picture Drama - Episode 3: Magic of Resurrection", 8, 13},
		{kurokoIMDb, "NG-Shuu 1", 1, 1},
		{kurokoIMDb, "NG-Shuu 2 1", 13, 13},
		{kurokoIMDb, "Would You Like to Chat?", 23, 23},
		{kurokoIMDb, "The Best Present(Episode 75.5)", 33, 40},
		{kurokoIMDb, "Last Game NG Shuu", 34, 38},
		{kurokoIMDb, "Let's Talk", 35, 0},
	} {
		got, ok := matchSpecial(c.title, c.episode, c.videos)
		if !ok {
			got = 0
		}
		if got != c.want {
			t.Errorf("matchSpecial(%q, %d) = %d, want %d", c.title, c.episode, got, c.want)
		}
	}
}

func TestSpecialIdentityRefusesFilesNumberedForAnotherSeasonOrEpisode(t *testing.T) {
	req := core.SearchRequest{Type: core.TypeEpisode, Title: "Show", Season: 0, Episode: 6}
	for _, c := range []struct {
		file  string
		asked int
		want  bool // stamped S00E06
	}{
		{"Doctor.Who.2005.6x00.Un.Canto.Di.Natale.ITA.ENG.BDMux.XviD", 9, false}, // another season's special
		{"Haikyuu!!.S00E08.1080p.Blu-Ray.x265", 13, false},                     // IMDb 8, asked 13
		{"Haikyuu!!.S00E13.1080p.Blu-Ray.x265", 13, true},                      // the asked IMDb number
		{"Show.S00E06.1080p", 0, true},                                         // not renumbered
		{"[Erai-raws] Kuroko no Basket 3rd Season - 25.5 END [1080p]", 40, true}, // half episode
		{"[Anime Time] Haikyuu!! OAD - Lev Appears!", 1, true},                 // named by title
		{"Haikyuu!! OVA 02 - vs. Akaten (2016) [1080p RUS]", 3, true},           // OVA count, no season
	} {
		info := specialIdentity(core.ReleaseInfo{Seasons: []int{1, 2}}, req, false, c.file, "pack", c.asked)
		got := len(info.Seasons) == 1 && info.Seasons[0] == 0 && len(info.Episodes) == 1 && info.Episodes[0] == 6
		if got != c.want {
			t.Errorf("%q asked %d: stamped=%v, want %v", c.file, c.asked, got, c.want)
		}
	}
}
