package resolver

import (
	"testing"

	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/releaseparse"
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

// "[A&C] One Piece S10E01 (0337)" has no year or id; its absolute number is
// the evidence that it is the anime and not the live-action namesake.
func TestJinxAbsoluteNumberIsSeriesIdentityEvidence(t *testing.T) {
	req := core.SearchRequest{Type: core.TypeEpisode, Title: "One Piece", Season: 10, Episode: 337, AbsoluteEpisode: 337, EpisodeTitle: "Plunging into the Devil’s Sea! The Mysterious Skeleton Floating in the Fog!"}
	name := "[A&C] One Piece S10E01 (0337) (WEB HEVC 1080p) [Multi-Audio-Subs] [F161A3B0]"
	if !episodeIdentityEvidence(core.Release{Name: name, Info: releaseparse.Parse(name)}, req) {
		t.Fatal("an exact absolute number is not identity evidence")
	}
	wrong := "[A&C] One Piece S10E02 (0338) (WEB HEVC 1080p)"
	if episodeIdentityEvidence(core.Release{Name: wrong, Info: releaseparse.Parse(wrong)}, req) {
		t.Fatal("another episode's absolute number counted")
	}
	short := core.SearchRequest{Type: core.TypeEpisode, Title: "One Piece", Season: 1, Episode: 1, AbsoluteEpisode: 1}
	n1 := "One Piece S01E01 (0001) 1080p"
	if episodeIdentityEvidence(core.Release{Name: n1, Info: releaseparse.Parse(n1)}, short) {
		t.Fatal("a one-digit absolute number counted as evidence")
	}
}
