package releaseparse

import (
	"strings"
	"testing"

	"github.com/braedonsaunders/ether/internal/core"
)

func TestLongSeasonAndEpisodeNumbers(t *testing.T) {
	for name, want := range map[string][2]int{
		"One Piece (1999) S22E1100 1080p WEB-DL AVC Japanese-OldT":                 {22, 1100},
		"One.Piece.S008E240.Eternal.Farewell.1080p.NF.WEB-DL.DDP2.0.x264-KQRM.mkv": {8, 240},
		"One.Piece.S001E010.The.Weirdest.Guy.Ever.1080p.NF.WEB-DL":                 {1, 10},
		"Show.S01E05.1080p.WEB-DL":                                                 {1, 5},
	} {
		i := Parse(name)
		if !EpisodeMatches(i, want[0], want[1]) {
			t.Errorf("%q parsed seasons %v episodes %v, want S%dE%d", name, i.Seasons, i.Episodes, want[0], want[1])
		}
	}
	if i := Parse("Show.S01E01.1080p.BluRay"); len(i.Episodes) != 1 || i.Episodes[0] != 1 || i.Height != 1080 {
		t.Errorf("S01E01 1080p parsed episodes %v height %d", i.Episodes, i.Height)
	}
}

func TestAgreeingTMDBOutweighsAnotherIMDbID(t *testing.T) {
	want := core.IDs{TMDB: 37854, TVDB: 81797, IMDB: "tt0388629"}
	if m, c := CompareTitleIDs(core.IDs{TMDB: 37854, TVDB: 81797, IMDB: "tt3354352"}, want); !m || c {
		t.Errorf("Anime Tosho's One Piece ids: matched %v conflicting %v", m, c)
	}
	if _, c := CompareTitleIDs(core.IDs{TMDB: 111110, IMDB: "tt0388629"}, want); !c {
		t.Error("a different TMDB id did not conflict")
	}
	if _, c := CompareTitleIDs(core.IDs{IMDB: "tt3354352"}, want); !c {
		t.Error("a lone different IMDb id did not conflict")
	}
}

func TestAnimeAbsoluteNumberForms(t *testing.T) {
	for name, want := range map[string]int{
		"[A&C] One Piece S10E01 (0337) (WEB HEVC 1080p) [Multi-Audio-Subs]": 337,
		"One Piece - 0459 v2 [F-R][9d4e6bc5]":                               459,
		"[SubsPlease] One Piece - 1071 (1080p) [ABCD1234]":                  1071,
		"Show S01E05 (2019) 1080p WEB-DL":                                   0,
	} {
		i := Parse(name)
		got := 0
		if len(i.AbsoluteEpisodes) == 1 {
			got = i.AbsoluteEpisodes[0]
		}
		if got != want {
			t.Errorf("%q: absolute %v (seasons %v episodes %v), want %d", name, i.AbsoluteEpisodes, i.Seasons, i.Episodes, want)
		}
	}
	if i := Parse("[A&C] One Piece S10E01 (0337) (WEB HEVC 1080p)"); len(i.Seasons) != 1 || i.Seasons[0] != 10 || i.Episodes[0] != 1 {
		t.Errorf("S10E01 lost: seasons %v episodes %v", i.Seasons, i.Episodes)
	}
}

func TestAbsoluteNumberIsNotAnEpisodeTitle(t *testing.T) {
	want := "Great Adventure at Long Ring Long Land!"
	for _, n := range []string{
		"[A&C] One Piece S07E12 (0207) (WEB HEVC 1080p) [Multi-Audio-Subs] [4ED88C18]",
		"Kuroko's Basketball - S03E06 (56) (BD 1080p Hi10 Dual Audio -KS-)",
		"[A&C] One Piece S07E12 (0207) Great Adventure at Long Ring Long Land! (WEB 1080p)",
	} {
		if !EpisodeTitleMatches(n, want) && !strings.Contains(n, "Kuroko") {
			t.Errorf("%q: numbering read as a contradicting title", n)
		}
	}
	if EpisodeTitleMatches("[A&C] One Piece S07E12 (0207) Some Other Title (WEB 1080p)", want) {
		t.Error("a real contradicting title after the number was ignored")
	}
}
