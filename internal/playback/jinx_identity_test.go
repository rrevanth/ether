package playback

import (
	"testing"

	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/releaseparse"
)

// The resolver accepts long anime by their absolute number; the playback
// recheck must agree (One Piece TMDB S07E207 is A&C's "S07E12 (0207)").
func TestPlaybackIdentityTakesAbsoluteNumbers(t *testing.T) {
	for name, want := range map[string]bool{
		"[A&C] One Piece S07E12 (0207) (WEB HEVC 1080p) [Multi-Audio-Subs]": true,
		"One Piece S01E207 1080p":       true,
		"One Piece S07E12 1080p":        false,
		"One Piece S07E13 (0208) 1080p": false,
	} {
		if got := episodeIdentity(releaseparse.Parse(name), 7, 207, 207); got != want {
			t.Errorf("%q: %v, want %v", name, got, want)
		}
	}
	title := &core.Title{Type: core.TypeEpisode, Season: 7, Episode: 207, AbsoluteEpisode: 207}
	if !payloadMatchesTitle([]byte(debridPayloadPrefix+`{"path":"One Piece/[A&C] One Piece S07E12 (0207) (WEB HEVC 1080p).mkv"}`), title) {
		t.Error("the opened A&C file was rejected")
	}
}

// A special's title is the resolver's to judge (partial titles agree); an
// episode of unknown number still gets upstream's title check.
func TestPlaybackLeavesSpecialTitlesToTheResolver(t *testing.T) {
	special := &core.Title{Type: core.TypeEpisode, Season: 0, Episode: 8, Title: "Episode of Luffy: Adventure on Hand Island"}
	if !releaseMatchesTitle(core.Release{Name: "One.Piece.S00E08.Episode.of.Luffy.CUSTOM.MULTi.1080p.BDRip.x264"}, special) {
		t.Error("the special was refused at open")
	}
	unknown := &core.Title{Type: core.TypeEpisode, ShowYear: 2005, Title: "Stone and Sky"}
	if releaseMatchesTitle(core.Release{Name: "Space.Pilot.S02E02.Other.Story.1080p.WEB"}, unknown) {
		t.Error("an episode with another title opened")
	}
}

func TestSpecialPayloadInACollectionFolderOpens(t *testing.T) {
	special := &core.Title{Type: core.TypeEpisode, Season: 0, Episode: 8, Title: "Episode of Luffy: Adventure on Hand Island"}
	ok := payloadMatchesTitle([]byte(debridPayloadPrefix+`{"path":"One Piece (0001-1071 Movies Specials) [BD CR]/Specials/[Anime Time] One Piece Special - Episode of Luffy.mkv"}`), special)
	imdb := payloadMatchesTitle([]byte(debridPayloadPrefix+`{"path":"One.Piece.S00E24.Episode.of.Luffy.1080p.mkv"}`), special)
	if !ok || !imdb {
		t.Fatalf("collection file %v, add-on-numbered file %v", ok, imdb)
	}
}
