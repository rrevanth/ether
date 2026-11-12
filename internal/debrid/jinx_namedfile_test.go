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
