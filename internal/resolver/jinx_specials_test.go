package resolver

import (
	"testing"

	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/releaseparse"
)

// Real One Piece special candidates (TMDB season 0), plus specials of other
// shows that already played correctly.
func TestSpecialsMustNameThemselvesAndAgree(t *testing.T) {
	sp := func(ep int, title string) core.SearchRequest {
		return core.SearchRequest{Type: core.TypeEpisode, Title: "One Piece", Season: 0, Episode: ep, EpisodeTitle: title}
	}
	for _, c := range []struct {
		name string
		req  core.SearchRequest
		want string
	}{
		{"One.Piece.S00E01.One.Piece.Heroines.1080p.CR.WEB-DL.DUAL", sp(1, "Contingency Plan. Complete Summary One Piece"), "special's title differs"},
		{"[A&C] One Piece - S00E02 - Movie 01 (BD 1080p HEVC)", sp(2, "Adventure in the Ocean's Navel"), "special's title differs"},
		{"One Piece - S00E06 - The Detective Memoirs of Chief Straw Hat Luffy", sp(6, "New Year Special: Special Report - Secret"), "special's title differs"},
		{"[Mega Pack] One.Piece.Ep.001-165.MULTi.1080p", sp(3, "Open Upon the Great Sea! A Father's Huge, HUGE Dream!"), "names no special"},
		{"One.Piece.S00E08.Episode.of.Luffy.CUSTOM.MULTi.1080p.BDRip.x264", sp(8, "Episode of Luffy: Adventure on Hand Island"), ""},
		{"One.Piece.S00E11.3D2Y.CUSTOM.MULTi.1080p.BDRip.x264", sp(11, "3D2Y: Overcome Ace's Death! Luffy's Vow to his Friends"), ""},
		{"One Piece - S00E04 - Protect! The Last Great Performance", sp(4, "Protect! The Last Great Performance"), ""},
		{"One Piece (0001-1071 Movies Specials) [BD CR] [Dual Audio]", sp(9, "Glorious Island"), ""},
		{"Kuroko's Basketball (2012) - S00E01 - NG-Shuu 1 [Bluray-1080p][Opus 2.0][AV1]-BlackRabbit", core.SearchRequest{Type: core.TypeEpisode, Season: 0, Episode: 1, EpisodeTitle: "NG-Shuu 1"}, ""},
		{"Doctor Who 2005 S00E04 The Runaway Bride 1080p HMAX WEB-DL DD 2 0 H 264-SLiGNOME", core.SearchRequest{Type: core.TypeEpisode, Season: 0, Episode: 4, EpisodeTitle: "The Runaway Bride"}, ""},
		{"Haikyuu!!.S00E04.1080p.Blu-Ray.10-Bit.Dual-Audio.DTS-HD.x265-iAHD", core.SearchRequest{Type: core.TypeEpisode, Season: 0, Episode: 4, EpisodeTitle: "Land vs. Air"}, ""},
	} {
		r := withParsed(rel("x", c.name, 1400))
		if got := rejectSpecial(r, c.req); got != c.want {
			t.Errorf("%q: %q, want %q (title %q)", c.name, got, c.want, releaseparse.SpecialReleaseTitle(c.name))
		}
	}
	// Regular episodes are untouched.
	if got := rejectSpecial(withParsed(rel("y", "[Mega Pack] One.Piece.Ep.001-165.MULTi.1080p", 1400)), core.SearchRequest{Type: core.TypeEpisode, Season: 1, Episode: 3}); got != "" {
		t.Errorf("regular episode rejected: %q", got)
	}
}
