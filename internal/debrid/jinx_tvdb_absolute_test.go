package debrid

import (
	"testing"

	"github.com/braedonsaunders/ether/internal/core"
)

// One Piece S23E1176 (TMDB) is "S01E1176" in TVDB-numbered releases: the same absolute number.
func TestJinxChooseFileTVDBAbsoluteNumbering(t *testing.T) {
	target := &core.ReleaseTarget{Season: 23, Episode: 1176, AbsoluteEpisode: 1176, ShowYear: 1999, ExpectedIDs: core.IDs{TMDB: 37854}}
	for _, tc := range []struct {
		path string
		want bool
	}{
		{"One.Piece.S01E1176.1080p.NF.WEB-DL.AAC2.0.H.264-VARYG.mkv", true},
		{"One.Piece.S01E1176.The.Threat.of.the.Arrow-Arrow.Fruit-Gunko.vs.the.Straw.Hats.REPACK.1080p.CR.WEB-DL.AAC2.0.H.264-VARYG.mkv", true},
		{"One.Piece.S23E1176.1080p.WEB-DL.mkv", true},
		{"One.Piece.S01E1175.1080p.NF.WEB-DL.AAC2.0.H.264-VARYG.mkv", false},
		{"One.Piece.S01E1177.1080p.NF.WEB-DL.AAC2.0.H.264-VARYG.mkv", false},
		{"One.Piece.S02E1176.1080p.WEB-DL.mkv", false},
	} {
		got, err := ChooseFile([]File{{Path: tc.path, Size: 1_400_000_000}}, target)
		if tc.want && (err != nil || got.Path != tc.path) {
			t.Errorf("%q should be chosen: %+v %v", tc.path, got, err)
		}
		if !tc.want && err == nil {
			t.Errorf("%q must not be chosen for S23E1176", tc.path)
		}
	}
	// A per-season show with a small absolute number is unaffected: S01E17 is not season 2's episode 5.
	short := &core.ReleaseTarget{Season: 2, Episode: 5, AbsoluteEpisode: 17}
	if got, err := ChooseFile([]File{{Path: "Show.S01E17.1080p.mkv", Size: 800_000_000}}, short); err == nil {
		t.Errorf("a short show's S01E17 was chosen for S02E05: %+v", got)
	}
}
