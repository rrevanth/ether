package releaseparse

import "testing"

func TestSpecialFileNamingAnotherSpecialOrOnlyACount(t *testing.T) {
	aot := []string{"Attack on Titan", "進撃の巨人", "Shingeki no Kyojin"}
	aotSiblings := []string{"Since That Day", "Ilse's Notebook: Memoirs of a Recon Corps Member",
		"The Sudden Visitor: The Torturous Curse of Youth", "Distress", "Lost Girls (1)", "Chibi Theater: Fly, Cadets, Fly!: Day 3"}
	for _, c := range []struct {
		path, want      string
		shows, siblings []string
		another, count  bool
	}{
		{"Attack on Titan Season 0/A Sudden Visitor The Torturous Curse of Adolescence/Shingeki no Kyojin OVA - 2 (s00e02).mkv",
			"Chibi Theater: Fly, Cadets, Fly!: Day 1", aot, aotSiblings, true, false},
		{"Attack on Titan Season 0/Distress/Shingeki no Kyojin OVA - 3 (s00e03).mkv", "Chibi Theater: Fly, Cadets, Fly!: Day 3", aot, aotSiblings, true, false},
		{"Attack on Titan Season 0/Lost Girls/Shingeki no Kyojin - Lost Girls OVA - 01 (s00e06).mkv", "Chibi Theater: Fly, Cadets, Fly!: Day 11", aot, aotSiblings, true, false},
		{"[Kawaiika-Raws] Shingeki no Kyojin OAD [DVDRip 1024x576 HEVC FLAC]/[Kawaiika-Raws] Shingeki no Kyojin OAD 02 [DVDRip 1024x576 HEVC FLAC].mkv",
			"Chibi Theater: Fly, Cadets, Fly!: Day 14", aot, aotSiblings, false, true},
		{"ZR/Shingeki no Kyojin S00E13 (Distress) 2014 1080p Bluray REMUX AVC AAC 2.0 Dual Audio -ZR-.mkv", "Distress", aot, aotSiblings, false, false},
		{"[Phr0stY]/Attack on Titan (Shingeki no Kyojin) Chibi Shorts  Day 05-07 (1920x1080).mkv", "Chibi Theater: Fly, Cadets, Fly!: Day 5", aot, aotSiblings, false, false},
		{"[Commie] Haikyuu!! Land vs. Sky [BD 720p AAC]/[Commie] Haikyuu!! Land vs. Air - 01 [BD 720p AAC] [A90E15CC].mkv",
			"Haikyu Quest Picture Drama - Episode 3", []string{"Haikyu!!", "ハイキュー!!", "Haikyuu!!"}, []string{"Land vs. Air", "The Path of the Ball"}, true, false},
		{"[Erai-raws] Kuroko no Basket - 25.5 [1080p][Multiple Subtitle]/[Erai-raws] Kuroko no Basket - 22.5 [1080p][Multiple Subtitle].mkv",
			"NG-Shuu 9", []string{"Kuroko's Basketball", "黒子のバスケ"}, []string{"NG-Shuu 1", "Tip Off"}, false, true},
		{"Doctor.Who.2005.S00E02.1080p.BluRay.AV1-PTNX/Doctor.Who.2005.S00E02.1080p.BluRay.AV1-PTNX.mkv", "The Christmas Invasion",
			[]string{"Doctor Who"}, []string{"The Runaway Bride"}, false, false},
		{"Doctor Who S03 E01-13 (+speciale) 1080p BluRay x265 HEVC 10bit Ac3 Ita_Eng Subs/S00E04 - La Sposa Perfetta.mkv", "The Runaway Bride",
			[]string{"Doctor Who"}, []string{"The Christmas Invasion"}, false, false},
		{"[Erai-raws] One Piece - Umareru Rekishi! Gekidou no Shinkyuu Yonkou! [1080p][HEVC][Multiple Subtitle][84E3A27E].mkv",
			"Making History! The Turbulent Old and New Four Emperors!", []string{"One Piece"}, []string{"Heart of Gold"}, false, false},
		{"[Judas] One Piece - Barto no Himitsu no Heya - SP03.mkv", "Luffy-senpai Support Project! Barto's Secret Room! 1",
			[]string{"One Piece"}, []string{"Barto's Secret Room!"}, false, false},
		{"One Piece Specials Collection V2/One.Piece.S00E024.Episode.of.Luffy.CUSTOM.MULTi.1080p.BDRip.x264-R1.mkv", "Episode of Luffy: Adventure on Hand Island",
			[]string{"One Piece"}, []string{"Episode of East Blue - Luffy and His 4 Crewmates' Big Adventure", "Episode of Nami: Tears of a Navigator"}, false, false},
	} {
		if got := SpecialNamesAnother(c.path, c.want, c.siblings, c.shows); got != c.another {
			t.Errorf("SpecialNamesAnother(%q, %q) = %v", c.path, c.want, got)
		}
		if got := SpecialCountOnly(c.path, c.shows); got != c.count {
			t.Errorf("SpecialCountOnly(%q) = %v", c.path, got)
		}
	}
}
