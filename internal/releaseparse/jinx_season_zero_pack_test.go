package releaseparse

import "testing"

func TestJinxSeasonZeroPackIsNotACount(t *testing.T) {
	aot := []string{"Attack on Titan", "Shingeki no Kyojin"}
	for _, n := range []string{"Attack.on.Titan.S00.Specials.1080p.BluRay.Remux.AVC.FLAC.2.0-Flugel", "[BlackRabbit] Attack on Titan (2013) - S00 [Bluray-1080p]"} {
		if SpecialCountOnly(n, aot...) {
			t.Errorf("%q: a season-zero pack counted as a bare OAD number", n)
		}
	}
	// Bare counts stay count-only.
	for _, n := range []string{"[Kawaiika-Raws] Shingeki no Kyojin OAD 02 [DVDRip 1024x576 HEVC FLAC]", "[Erai-raws] Shingeki no Kyojin OAD - 02 [1080p]"} {
		if !SpecialCountOnly(n, aot...) {
			t.Errorf("%q: a bare OAD count accepted", n)
		}
	}
}
