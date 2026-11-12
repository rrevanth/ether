package resolver

import (
	"context"
	"testing"
	"time"

	"github.com/braedonsaunders/ether/internal/core"
)

// The 2026 animated series' TMDB alternative title is plain "RAMAYAN"; a
// yearless "Ramayan.S01E01-44" belongs to the 1987 Ramayan the catalogue
// also holds. Its own full-title release still matches.
func TestAlternativeTitleNamingAnotherShowIsRejected(t *testing.T) {
	asked := map[string]int{}
	s := New(&fakeIndexer{}, fakeHealth{}, nil, testCfg(), Options{TitleNamesake: func(_ context.Context, _ core.RatingKey, title string) ([]Namesake, bool, error) {
		asked[title]++
		if title == "Ramayan" {
			return []Namesake{{Key: 2000000000001990, Year: 1987}}, false, nil
		}
		return nil, true, nil
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
	if asked["Ramayan"] != 1 || asked["RAMAYAN AN ATHARV SRIVASTAVA SERIES"] != 1 || len(asked) != 2 {
		t.Fatalf("catalogue asked %v, want once each for Ramayan and the show's own title", asked)
	}
}

// An alternative title that names no other show (a romanized original)
// keeps matching.
func TestAlternativeTitleWithoutNamesakeStillMatches(t *testing.T) {
	s := New(&fakeIndexer{}, fakeHealth{}, nil, testCfg(), Options{TitleNamesake: func(context.Context, core.RatingKey, string) ([]Namesake, bool, error) { return nil, true, nil }})
	title := &core.Title{Type: core.TypeEpisode, ShowKey: 2000000000209867, ShowTitle: "Frieren: Beyond Journey's End", ShowAlternativeTitles: []string{"Sousou no Frieren"}, Season: 1, Episode: 3}
	sreq := core.SearchRequest{Type: core.TypeEpisode, Title: title.ShowTitle, Season: 1, Episode: 3, AlternativeTitles: title.ShowAlternativeTitles}
	r := withParsed(rel("f", "[SubsPlease] Sousou no Frieren - 03 (1080p) [ABCD1234].mkv", 1400))
	s.prepareNamesakes(context.Background(), title, sreq, []core.Release{r})
	if got := s.rejectNamesake(r, title, sreq); got != "" {
		t.Fatalf("romanized title rejected: %q", got)
	}
}

// Two shows titled "Ramayan": an unyeared release belongs to the 1987 one
// (far more votes), so the 2008 show needs a year in the name.
func TestSharedTitleUnyearedReleaseGoesToTheBetterKnownShow(t *testing.T) {
	owner := map[core.RatingKey]bool{2000000000001990: true, 2000000000015620: false}
	s := New(&fakeIndexer{}, fakeHealth{}, nil, testCfg(), Options{TitleNamesake: func(_ context.Context, show core.RatingKey, _ string) ([]Namesake, bool, error) {
		other := core.RatingKey(2000000000001990)
		if show == other {
			other = 2000000000015620
		}
		return []Namesake{{Key: other}}, owner[show], nil
	}})
	pack := withParsed(rel("t", "Ramayan.S01E01.Birth.of.the.Four.Princes.1080p.DSNP+.WEB-DL.AAC2.0.x264-Telly", 1400))
	yeared := withParsed(rel("y", "Ramayan (2008) S01E01 720p WEB-DL x264 AAC 2.0-JA", 700))
	for show, wantPack := range map[core.RatingKey]string{2000000000001990: "", 2000000000015620: "unyeared title belongs to a better-known show"} {
		title := &core.Title{Type: core.TypeEpisode, ShowKey: show, ShowTitle: "Ramayan", Season: 1, Episode: 1}
		sreq := core.SearchRequest{Type: core.TypeEpisode, Title: "Ramayan", Season: 1, Episode: 1}
		s.prepareNamesakes(context.Background(), title, sreq, []core.Release{pack, yeared})
		if got := s.rejectNamesake(pack, title, sreq); got != wantPack {
			t.Errorf("show %d: unyeared pack %q, want %q", show, got, wantPack)
		}
		if got := s.rejectNamesake(yeared, title, sreq); got != "" {
			t.Errorf("show %d: yeared release rejected: %q", show, got)
		}
	}
}

// "One Piece" is the 1999 anime (5,645 votes) and the 2023 live action
// (1,997): the anime owns unyeared releases, but not ones carrying the live
// action's year or its title for the same episode.
func TestOwnerTurnsAwayReleasesNamingTheOtherShow(t *testing.T) {
	live := core.RatingKey(2000000000111110)
	s := New(&fakeIndexer{}, fakeHealth{}, nil, testCfg(), Options{
		TitleNamesake: func(context.Context, core.RatingKey, string) ([]Namesake, bool, error) {
			return []Namesake{{Key: live, Year: 2023}}, true, nil
		},
		NamesakeEpisodeTitle: func(_ context.Context, show core.RatingKey, season, episode int) (string, error) {
			if show == live && season == 1 && episode == 1 {
				return "ROMANCE DAWN", nil
			}
			return "", nil
		},
	})
	title := &core.Title{Type: core.TypeEpisode, ShowKey: 2000000000037854, ShowTitle: "One Piece", ShowYear: 1999, Season: 1, Episode: 1}
	sreq := core.SearchRequest{Type: core.TypeEpisode, Title: "One Piece", Year: 1999, Season: 1, Episode: 1, EpisodeTitle: "I'm Luffy! The Man Who Will Become the Pirate King!"}
	aired2023 := sreq
	aired2023.Released = time.Date(2023, 10, 29, 0, 0, 0, 0, time.UTC)
	if r := withParsed(rel("y", "[SubsPlease] One Piece - 1080 (2023) (1080p)", 1400)); s.rejectNamesake(r, title, aired2023) != "" {
		t.Error("an episode's own air year read as the live action's")
	}
	for name, want := range map[string]string{
		"ONE.PIECE.S01E01.ROMANCE.DAWN.2160p.NF.WEB-DL.DDP5.1.Atmos.DV.HDR.H.265-G66":            "names another show's episode",
		"One Piece S01E01 2023 2160p NF WEB-DL DDP5 1 Atmos DV HDR H 265-HHWEB":                  "names another show's year",
		"[Erai-raws] One Piece - 001 [1080p][Multiple Subtitle]":                                 "",
		"One.Piece.S01E01.Im.Luffy.The.Man.Who.Will.Become.the.Pirate.King.1080p.NF.WEB-DL.x264": "",
	} {
		r := withParsed(rel("x", name, 1400))
		s.prepareNamesakes(context.Background(), title, sreq, []core.Release{r})
		if got := s.rejectNamesake(r, title, sreq); got != want {
			t.Errorf("%q: %q, want %q", name, got, want)
		}
	}
}

// One Piece S01E01 exists in the 1999 anime and the 2023 live action: a
// release saying only "One.Piece.S01E01" might be either and ranks after
// every release that names the anime; past S01E08 (no live-action episode)
// nothing is ambiguous.
func TestAmbiguousSharedTitleReleasesRankLast(t *testing.T) {
	live := core.RatingKey(2000000000111110)
	s := New(&fakeIndexer{}, fakeHealth{}, nil, testCfg(), Options{
		TitleNamesake: func(context.Context, core.RatingKey, string) ([]Namesake, bool, error) {
			return []Namesake{{Key: live, Year: 2023}}, true, nil
		},
		NamesakeEpisodeTitle: func(_ context.Context, show core.RatingKey, season, episode int) (string, error) {
			if show == live && season == 1 && episode <= 8 {
				return "ROMANCE DAWN", nil
			}
			return "", nil
		},
	})
	title := &core.Title{Type: core.TypeEpisode, ShowKey: 2000000000037854, ShowTitle: "One Piece", ShowYear: 1999, Season: 1, Episode: 1}
	sreq := core.SearchRequest{Type: core.TypeEpisode, Title: "One Piece", Year: 1999, Season: 1, Episode: 1, EpisodeTitle: "I'm Luffy! The Man Who Will Become the Pirate King!", Released: time.Date(1999, 10, 20, 0, 0, 0, 0, time.UTC)}
	names := map[string]bool{
		"One.Piece.S01E01.2160p.NF.WEB-DL.DDP5.1.Atmos.DV.HDR-RHS.mkv":                           true,
		"[Feibanyama] One Piece EP0001 [IQIYI WebRip 2160p HEVC OPUS]":                           false,
		"One Piece - 001 1080p WEB x264 -NanDesuKa (CR)":                                         false,
		"One.Piece.S01E01.Im.Luffy.The.Man.Who.Will.Become.the.Pirate.King.1080p.NF.WEB-DL.x264": false,
		"One Piece (1999) S01E01 1080p WEB-DL":                                                   false,
	}
	var rels []core.Release
	for n := range names {
		rels = append(rels, withParsed(rel(n, n, 1400)))
	}
	s.prepareNamesakes(context.Background(), title, sreq, rels)
	for _, r := range rels {
		if got := s.namesakeAmbiguous(r, title, sreq); got != names[r.Name] {
			t.Errorf("%q: ambiguous %v, want %v", r.Name, got, names[r.Name])
		}
	}
	later := sreq
	later.Episode, later.EpisodeTitle = 9, "Honorable Liar? Captain Usopp!"
	r9 := withParsed(rel("e9", "One.Piece.S01E09.1080p.NF.WEB-DL.x264", 1400))
	s.prepareNamesakes(context.Background(), title, later, []core.Release{r9})
	if s.namesakeAmbiguous(r9, title, later) {
		t.Error("S01E09 has no live-action episode but was ambiguous")
	}
	if !compareCandidates(candidate{langOK: true}, candidate{langOK: true, ambiguous: true, keys: rankKeys{position: 99}}) {
		t.Error("an ambiguous 2160p release outranked a release naming the show")
	}
}

func TestPlayVersionsLeaveOutAmbiguousReleases(t *testing.T) {
	a, b, c := candidate{rel: core.Release{ID: "anime"}}, candidate{rel: core.Release{ID: "live?"}, ambiguous: true}, candidate{rel: core.Release{ID: "anime-1080"}}
	if got := withoutAmbiguous([]candidate{a, b, c}); len(got) != 2 || got[0].rel.ID != "anime" || got[1].rel.ID != "anime-1080" {
		t.Fatalf("versions %+v", got)
	}
	if got := withoutAmbiguous([]candidate{b}); len(got) != 1 {
		t.Fatal("the only release was dropped")
	}
}
