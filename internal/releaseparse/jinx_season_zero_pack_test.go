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

func TestJinxSPAndZeroXNameSpecials(t *testing.T) {
	for _, n := range []string{"[Judas] One Piece - SP30 - A Comprehensive Anatomy! The Legend of Kozuki Oden!.mkv", "[SubsPlease] Jujutsu Kaisen - SP1 (1080p)",
		"[Erai-raws] Jujutsu Kaisen - SP 01 [1080p]", "Jujutsu.Kaisen.S02SP1.1080p.WEB-DL", "One Piece - 0x42 - A Comprehensive Anatomy!.mkv"} {
		if !NamesSpecial(n) {
			t.Errorf("%q does not name a special", n)
		}
	}
	for _, n := range []string{"Spider-Man.2002.1080p.BluRay", "One.Piece.S01E01.1080p.WEB"} {
		if NamesSpecial(n) {
			t.Errorf("%q names a special", n)
		}
	}
}
