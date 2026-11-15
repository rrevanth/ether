package resolver

import (
	"testing"

	"github.com/braedonsaunders/ether/internal/core"
)

// One Piece (1999) has a same-named live-action series; its specials are
// told apart by their own titles, not by an S00Exx number they rarely carry.
func TestJinxSpecialTitleIsSeriesIdentityEvidence(t *testing.T) {
	req := core.SearchRequest{Type: core.TypeEpisode, Title: "One Piece", Season: 0, Episode: 19, EpisodeTitle: "A Comprehensive Anatomy! The Legend of Kozuki Oden!"}
	good := core.Release{Name: "[Judas] One Piece - SP30 - A Comprehensive Anatomy! The Legend of Kozuki Oden! [1080p][HEVC x265 10bit][Multi-Subs] (Weekly)"}
	if !episodeIdentityEvidence(good, req) {
		t.Fatal("the special's own title is not identity evidence")
	}
	other := core.Release{Name: "[Judas] One Piece - Barto no Himitsu no Heya - SP03.mkv"}
	if episodeIdentityEvidence(other, req) {
		t.Fatal("another special's name counted as evidence")
	}
	regular := core.SearchRequest{Type: core.TypeEpisode, Title: "One Piece", Season: 1, Episode: 1, EpisodeTitle: "I'm Luffy! The Man Who Will Become the Pirate King!"}
	if episodeIdentityEvidence(core.Release{Name: "[HorribleSubs] One Piece - 01 [1080p]"}, regular) {
		t.Fatal("a name-only regular episode gained evidence")
	}
}
