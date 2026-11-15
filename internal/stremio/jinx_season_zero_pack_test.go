package stremio

import (
	"testing"

	"github.com/braedonsaunders/ether/internal/core"
)

// Torrentio answers Kuroko's Basketball tt2348803:0:1 with a file from a
// pack labelled only "S00"; that pack is a specials pack.
func TestJinxSeasonZeroPackSuppliesItsSpecial(t *testing.T) {
	torrent := "[BlackRabbit] Kuroko's Basketball (2012) - S00 [Bluray-1080p][Opus 2.0][AV1]"
	req := core.SearchRequest{Type: core.TypeEpisode, Title: "Kuroko's Basketball", Season: 0, Episode: 1, EpisodeTitle: "NG-Shuu 1"}
	file := "Kuroko's Basketball (2012) - S00E01 - NG-Shuu 1 [Bluray-1080p][Opus 2.0][AV1]-BlackRabbit.mkv"
	rel, keep := specialRelease(core.Release{Name: torrent}, req, file, torrent, 1)
	if !keep || rel.Target == nil || rel.Target.Episode != 1 {
		t.Fatalf("season-zero pack's special dropped: keep=%v", keep)
	}
	// Still refused: another special's file from the same pack, and a regular-season pack.
	if _, keep := specialRelease(core.Release{Name: torrent}, req, "Kuroko's Basketball (2012) - S00E02 - NG-Shuu 2 [Bluray-1080p]-BlackRabbit.mkv", torrent, 1); keep {
		t.Fatal("S00E02 accepted for S00E01")
	}
	s2 := "[BlackRabbit] Kuroko's Basketball (2012) - S02 [Bluray-1080p][Opus 2.0][AV1]"
	if _, keep := specialRelease(core.Release{Name: s2}, req, file, s2, 1); keep {
		t.Fatal("a regular-season pack supplied a special")
	}
}
