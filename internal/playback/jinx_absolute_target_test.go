package playback

import (
	"encoding/json"
	"testing"

	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/releaseparse"
)

// One Piece S10E337 from "[A&C] ... Season 10 [337-381]": the file names
// the in-season number and the absolute one ("S10E01 (0337)"). Playback's
// title can lack the absolute number; the resolver's target carries it.
func TestJinxArcSeasonFileUsesTheTargetAbsolute(t *testing.T) {
	name := "[A&C] One Piece S10E01 (0337) (WEB HEVC 1080p) [Multi-Audio-Subs] [F161A3B0]"
	path := "[A&C] One Piece Season 10 - Thriller Bark [337-381] (WEB) [Multi-Audio-Subs]/" + name + ".mkv"
	title := &core.Title{Type: core.TypeEpisode, Season: 10, Episode: 337, ShowTitle: "One Piece"}
	target := &core.ReleaseTarget{Season: 10, Episode: 337, AbsoluteEpisode: 337}
	b, _ := json.Marshal(map[string]any{"path": path, "target": target})
	if !payloadMatchesTitle(append([]byte(debridPayloadPrefix), b...), title) {
		t.Fatal("payload with the target's absolute number rejected")
	}
	if !releaseMatchesTitle(core.Release{Name: name, Info: releaseparse.Parse(name), Target: target}, title) {
		t.Fatal("release with the target's absolute number rejected")
	}
	// The fallback is only the target's own episode, and a wrong number still fails.
	other := &core.ReleaseTarget{Season: 10, Episode: 338, AbsoluteEpisode: 337}
	if releaseMatchesTitle(core.Release{Name: name, Info: releaseparse.Parse(name), Target: other}, title) {
		t.Fatal("another episode's target vouched for this one")
	}
	wrong := "[A&C] One Piece S10E02 (0338) (WEB HEVC 1080p)"
	if releaseMatchesTitle(core.Release{Name: wrong, Info: releaseparse.Parse(wrong), Target: target}, title) {
		t.Fatal("episode 338 accepted for 337")
	}
}
