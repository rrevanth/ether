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
		"One Piece S01E207 1080p":                                           true,
		"One Piece S07E12 1080p":                                            false,
		"One Piece S07E13 (0208) 1080p":                                     false,
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
