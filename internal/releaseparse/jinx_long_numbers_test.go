package releaseparse

import "testing"

func TestLongSeasonAndEpisodeNumbers(t *testing.T) {
	for name, want := range map[string][2]int{
		"One Piece (1999) S22E1100 1080p WEB-DL AVC Japanese-OldT":                 {22, 1100},
		"One.Piece.S008E240.Eternal.Farewell.1080p.NF.WEB-DL.DDP2.0.x264-KQRM.mkv": {8, 240},
		"One.Piece.S001E010.The.Weirdest.Guy.Ever.1080p.NF.WEB-DL":                 {1, 10},
		"Show.S01E05.1080p.WEB-DL":                                                 {1, 5},
	} {
		i := Parse(name)
		if !EpisodeMatches(i, want[0], want[1]) {
			t.Errorf("%q parsed seasons %v episodes %v, want S%dE%d", name, i.Seasons, i.Episodes, want[0], want[1])
		}
	}
	if i := Parse("Show.S01E01.1080p.BluRay"); len(i.Episodes) != 1 || i.Episodes[0] != 1 || i.Height != 1080 {
		t.Errorf("S01E01 1080p parsed episodes %v height %d", i.Episodes, i.Height)
	}
}
