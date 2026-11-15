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
	rel, keep := specialRelease(core.Release{Name: torrent}, req, file, torrent, 1, true)
	if !keep || rel.Target == nil || rel.Target.Episode != 1 {
		t.Fatalf("season-zero pack's special dropped: keep=%v", keep)
	}
	// Still refused: another special's file from the same pack, and a regular-season pack.
	if _, keep := specialRelease(core.Release{Name: torrent}, req, "Kuroko's Basketball (2012) - S00E02 - NG-Shuu 2 [Bluray-1080p]-BlackRabbit.mkv", torrent, 1, true); keep {
		t.Fatal("S00E02 accepted for S00E01")
	}
	s2 := "[BlackRabbit] Kuroko's Basketball (2012) - S02 [Bluray-1080p][Opus 2.0][AV1]"
	if _, keep := specialRelease(core.Release{Name: s2}, req, file, s2, 1, true); keep {
		t.Fatal("a regular-season pack supplied a special")
	}
}

// Torrentio answers Haikyu!! IMDb S00E05 (TMDB S00E03 "Special Feature! The
// Spring Tournament of Their Youth") with romaji-titled OVA files from packs.
func TestJinxRomajiOVAFileTheAddonVouchedFor(t *testing.T) {
	req := core.SearchRequest{Type: core.TypeEpisode, Title: "Haikyu!!", AlternativeTitles: []string{"Haikyuu!!"}, Season: 0, Episode: 3,
		EpisodeTitle: "Special Feature! The Spring Tournament of Their Youth", Anime: true}
	torrent := "[DB] Haikyuu!! (Season 1-4+OVAs+Movies) [Dual Audio 10bit BD1080p][HEVC-x265]"
	file := "[DB]Haikyuu!! Tokushuu! Haru-kou Volley ni Kaketa Seishun_-_OVA_(10bit_WEB1080p_x265).mkv"
	rel, keep := specialRelease(core.Release{Name: torrent}, req, file, torrent, 5, true)
	if !keep || rel.Target == nil || rel.Target.AddonSpecialEpisode != 5 {
		t.Fatalf("romaji OVA the add-on answered for IMDb 5 dropped: keep=%v", keep)
	}
	at := "[Anime Time] Haikyuu!! Tokushuu! Haru Kou Volley ni Kaketa Seishun - OVA"
	if rel, keep := specialRelease(core.Release{Name: at}, req, "", at, 5, true); !keep || rel.Info.Title != "Haikyuu!!" {
		t.Fatalf("vouched romaji OVA's series title = %q (keep=%v), want the show's name", rel.Info.Title, keep)
	}
	// Not asked by a mapped IMDb number: no vouching.
	if _, keep := specialRelease(core.Release{Name: torrent}, req, file, torrent, 0, true); keep {
		t.Fatal("romaji OVA kept without the add-on's mapped answer")
	}
	// A bare group-local count with no title is still not taken on the add-on's word.
	if _, keep := specialRelease(core.Release{Name: torrent}, req, "[Kawaiika-Raws] Haikyuu!! OAD 02 [DVDRip 1024x576 HEVC FLAC].mkv", torrent, 5, true); keep {
		t.Fatal("a bare OAD count was taken on the add-on's word")
	}
	// An unlabelled file is not vouched for either.
	if _, keep := specialRelease(core.Release{Name: torrent}, req, "[DB]Haikyuu!! Tokushuu! Haru-kou Volley ni Kaketa Seishun.mkv", torrent, 5, true); keep {
		t.Fatal("an unlabelled file was taken on the add-on's word")
	}
}

// Torrentio's answer for One Piece IMDb S00E42 mixes two romaji-titled
// specials; nothing in it may be taken on the add-on's word.
func TestJinxDisagreeingAddonAnswerIsNotTrusted(t *testing.T) {
	req := core.SearchRequest{Type: core.TypeEpisode, Title: "One Piece", Season: 0, Episode: 19, EpisodeTitle: "A Comprehensive Anatomy! The Legend of Kozuki Oden!", Anime: true}
	mixed := []string{
		"[Judas] One Piece - SP30 - A Comprehensive Anatomy! The Legend of Kozuki Oden!.mkv",
		"[Erai-raws] One Piece - Momonosuke no Mei Shougun e no Michi - SP [1080p][Multiple Subtitle][0C08C78D].mkv",
		"[Erai-raws] One Piece - Dai Tettei Kaibou! Kouzuki Oden Densetsu! [1080p][Multiple Subtitle].mkv",
		"One Piece - 0x42 - A Comprehensive Anatomy! The Legend of Kozuki Oden!.mkv",
	}
	if addonAnswerAgrees(mixed, req) {
		t.Fatal("an answer naming two unverifiable specials was trusted")
	}
	one := []string{
		"[DB]Haikyuu!! Tokushuu! Haru-kou Volley ni Kaketa Seishun_-_OVA_(10bit_WEB1080p_x265).mkv",
		"[Erai-raws] Haikyuu!! Tokushuu! Haru Kou Volley ni Kaketa Seishun - OVA [1080p].mkv",
		"[Poopoo] Haikyu S00E05 (WEB 1080p HEVC AC3).mkv",
	}
	hq := core.SearchRequest{Type: core.TypeEpisode, Title: "Haikyu!!", AlternativeTitles: []string{"Haikyuu!!"}, Season: 0, Episode: 3, EpisodeTitle: "Special Feature! The Spring Tournament of Their Youth"}
	if !addonAnswerAgrees(one, hq) {
		t.Fatal("an answer whose unverifiable files all name one special was not trusted")
	}
}
