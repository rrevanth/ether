package debrid

import (
	"fmt"
	"testing"
	"time"

	"github.com/braedonsaunders/ether/internal/core"
)

// An add-on asked for S00E10 named the pack's file; files named by title
// alone ("Tip Off") are otherwise never chosen.
func TestChooseFilePrefersTheSourceNamedFile(t *testing.T) {
	files := []File{
		{Path: "Kuroko Movies/[Anime Time] Kuroko no Basket - Last Game.mkv", Size: 9 << 30},
		{Path: "Kuroko OVA/[Anime Time] Kuroko no Basket - Tip Off.mkv", Size: 1 << 30},
		{Path: "Kuroko OVA/[Anime Time] Kuroko no Basket - Tip Off.nfo", Size: 1 << 10},
	}
	target := &core.ReleaseTarget{Season: 0, Episode: 10, FileName: "[Anime Time] Kuroko no Basket - Tip Off"}
	f, err := ChooseFile(files, target)
	if err != nil || f.Path != files[1].Path {
		t.Fatalf("ChooseFile = %q, %v; want the named OVA", f.Path, err)
	}
}

// A hint that names no file, or two, falls back to the usual parsing.
func TestChooseFileIgnoresAnUnmatchedOrAmbiguousName(t *testing.T) {
	files := []File{
		{Path: "Show.S01/Show.S01E01.mkv", Size: 2 << 30},
		{Path: "Show.S01/Show.S01E02.mkv", Size: 2 << 30},
		{Path: "Show.S01/extra/Show.S01E02.mkv", Size: 1 << 30},
	}
	for _, name := range []string{"Something Else", "Show.S01E02"} {
		f, err := ChooseFile(files, &core.ReleaseTarget{Season: 1, Episode: 2, FileName: name})
		if err != nil || f.Path != files[1].Path {
			t.Fatalf("hint %q: ChooseFile = %q, %v; want the parsed S01E02", name, f.Path, err)
		}
	}
}

// A torrent whose files are another series is the release's fault: three
// of them in a row must not pause the account for everyone.
func TestDifferentTitleDoesNotPauseTheProvider(t *testing.T) {
	e := &entry{}
	now := time.Now()
	for i := 0; i < 5; i++ {
		if ev := e.observe(now, fmt.Errorf("Premiumize: %w", errDifferentTitle)); ev != nil && ev.State != "" {
			t.Fatalf("paused after %d mismatches: %+v", i+1, ev)
		}
	}
	if e.state != "" {
		t.Fatalf("state %q after title mismatches", e.state)
	}
}

// Torrentio numbers One Piece by IMDb: asked for TMDB S13E450 it named
// episode 422's file. A named file numbering another episode is ignored.
func TestChooseFileIgnoresANamedFileOfAnotherEpisode(t *testing.T) {
	files := []File{
		{Path: "One Piece/[Koten_Gars] One Piece - 0422 [1080p].mkv", Size: 1 << 30},
		{Path: "One Piece/[Koten_Gars] One Piece - 0450 [1080p].mkv", Size: 1 << 30},
	}
	f, err := ChooseFile(files, &core.ReleaseTarget{Season: 13, Episode: 450, AbsoluteEpisode: 450, FileName: "[Koten_Gars] One Piece - 0422 [1080p]"})
	if err != nil || f.Path != files[1].Path {
		t.Fatalf("ChooseFile = %q, %v; want episode 450", f.Path, err)
	}
}

// A pack of regular episodes holds no special: "005 - ..." is episode 5.
func TestChooseFileFindsNoSpecialInARegularPack(t *testing.T) {
	files := []File{
		{Path: "One Piece/001 - I'm Luffy.mkv", Size: 1 << 30},
		{Path: "One Piece/005 - Fear, Mysterious Power.mkv", Size: 1 << 30},
	}
	if f, err := ChooseFile(files, &core.ReleaseTarget{Season: 0, Episode: 5}); err == nil {
		t.Fatalf("S00E05 opened %q", f.Path)
	}
	if f, err := ChooseFile(append(files, File{Path: "One Piece/Specials/One Piece - S00E05 - Special.mkv", Size: 1 << 29}), &core.ReleaseTarget{Season: 0, Episode: 5}); err != nil || f.Path != "One Piece/Specials/One Piece - S00E05 - Special.mkv" {
		t.Fatalf("explicit S00E05 file: %q, %v", f.Path, err)
	}
}

// A special's pack file named with a partial title is the special.
func TestSpecialFileWithPartialTitlePasses(t *testing.T) {
	f := File{Path: "One Piece/Specials/One.Piece.S00E08.Episode.of.Luffy.CUSTOM.MULTi.1080p.BDRip.x264.mkv", Size: 1 << 30}
	target := &core.ReleaseTarget{Season: 0, Episode: 8, EpisodeTitle: "Episode of Luffy: Adventure on Hand Island"}
	if got, err := ChooseFile([]File{f}, target); err != nil || got.Path != f.Path {
		t.Fatalf("ChooseFile = %q, %v", got.Path, err)
	}
	wrong := File{Path: "One Piece/Specials/One Piece - S00E08 - Soccer King of Dreams.mkv", Size: 1 << 30}
	if _, err := ChooseFile([]File{wrong}, target); err == nil {
		t.Fatal("a file naming another special was chosen")
	}
}

// "One Piece - S00E01 ... (FUNi)" whose only file is "One Piece - 01": no special.
func TestSpecialRefusesARegularNumberedFile(t *testing.T) {
	files := []File{{Path: "One Piece - 01 - 1080p WEB H.264 -NanDesuKa (FUNi).mkv", Size: 1 << 30}}
	if f, err := ChooseFile(files, &core.ReleaseTarget{Season: 0, Episode: 1}); err == nil {
		t.Fatalf("opened %q for S00E01", f.Path)
	}
	ok := []File{{Path: "One Piece/Specials/One Piece - 01 - Episode of Nami.mkv", Size: 1 << 30}}
	if _, err := ChooseFile(ok, &core.ReleaseTarget{Season: 0, Episode: 1}); err == errNoVideo {
		t.Fatal("a file in a Specials folder was dropped")
	}
}

// IMDb numbers One Piece's "Detective Memoirs" special 14, TMDB 5: the
// add-on's named file opens for TMDB's S00E05.
func TestSpecialNamedFileKeepsTheAddonsNumbering(t *testing.T) {
	files := []File{{Path: "One.Piece.S00E14.The.Detective.Memoirs.of.Chief.Straw.Hat.Luffy.1080p.mkv", Size: 1 << 30}}
	f, err := ChooseFile(files, &core.ReleaseTarget{Season: 0, Episode: 5, FileName: "One.Piece.S00E14.The.Detective.Memoirs.of.Chief.Straw.Hat.Luffy.1080p"})
	if err != nil || f.Path != files[0].Path {
		t.Fatalf("ChooseFile = %q, %v", f.Path, err)
	}
}

// Asked for Jujutsu Kaisen's "Movie & Season 2 Pre-Broadcast Special", the
// add-on named BlackRabbit's TVDB-numbered S00E02, the film: its series
// year proves the show, not the special, and its title disagrees.
func TestSpecialNamedFileWithTheShowsYearStillNeedsItsTitle(t *testing.T) {
	shows := []string{"JUJUTSU KAISEN", "呪術廻戦", "Jujutsu Kaisen"}
	film := "JUJUTSU KAISEN (2020) - S00E02 - Jujutsu Kaisen 0 [Bluray-1080p][Opus 5.1][AV1]-BlackRabbit"
	files := []File{{Path: "[BlackRabbit] JUJUTSU KAISEN (2020) - S00 [Bluray-1080p][Opus 5.1][Dual Audio][AV1]/" + film + ".mkv", Size: 4 << 30}}
	target := &core.ReleaseTarget{Season: 0, Episode: 2, ShowYear: 2020, EpisodeTitle: "Movie & Season 2 Pre-Broadcast Special", FileName: film, ShowTitles: shows}
	if f, err := ChooseFile(files, target); err == nil {
		t.Fatalf("the film opened for the pre-broadcast special: %q", f.Path)
	}
	for _, c := range []struct{ name, title string }{
		{"Kuroko's Basketball (2012) - S00E11 - NG-Shuu 9 [Bluray-1080p][Opus 2.0][AV1]-BlackRabbit", "NG-Shuu 9"},
		{"Doctor Who (2005) - s00e02 - The Christmas Invasion - HDTV-1080p - h265", "The Christmas Invasion"},
	} {
		files := []File{{Path: "pack/" + c.name + ".mkv", Size: 1 << 30}}
		tg := &core.ReleaseTarget{Season: 0, Episode: 2, EpisodeTitle: c.title, FileName: c.name, ShowTitles: []string{"Kuroko's Basketball", "Doctor Who"}}
		if f, err := ChooseFile(files, tg); err != nil || f.Path != files[0].Path {
			t.Errorf("%q: ChooseFile = %q, %v", c.name, f.Path, err)
		}
	}
}
