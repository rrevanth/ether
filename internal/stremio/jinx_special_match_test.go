package stremio

import (
	"testing"

	"github.com/braedonsaunders/ether/internal/core"
)

func sv(ep int, name string) specialVideo { return specialVideo{Season: 0, Episode: ep, Name: name} }

// Cinemeta's (IMDb's) season 0 for Haikyu!! and Kuroko's Basketball, as listed on 2026-11-11.
var jinxHaikyuIMDb = []specialVideo{sv(1, "The Arrival of Lev"), sv(2, "Haikyu!! the Movie 1: The End and the Beginning"), sv(3, "Vs. Failing Grades"),
	sv(4, "Haikyu!! the Movie 2: The Winner and the Loser"), sv(5, "Special Feature! Betting on the Spring High Volleyball"), sv(6, "Haikyu!! the Movie 3: Talent and Sense"),
	sv(7, "Haikyu!! the Movie 4: Battle of Concepts"), sv(8, "Land vs. Air"), sv(9, "The Path of the Ball"), sv(11, "Haikyu!! Quest Picture Drama 01"),
	sv(12, "Haikyu!! Quest Picture Drama 02"), sv(13, "Haikyu!! Quest Picture Drama 03")}

var jinxKurokoIMDb = []specialVideo{sv(1, "NG-Shuu 1"), sv(2, "NG-Shuu 2"), sv(7, "Let's Chat OVA"), sv(12, "Idiots Can't Win!"), sv(13, "NG-Shuu 2 1"),
	sv(23, "Would You Like To Chat? OVA"), sv(24, "NG-Shuu 3 1"), sv(37, "Kuroko's Basketball: Last Game"), sv(38, "NG-Shuu Last Game"),
	sv(39, "Shall We Chat? OVA"), sv(40, "The Best Present (Episode 75.5)")}

func TestJinxMapSpecialRenumbersTMDBToIMDb(t *testing.T) {
	for _, c := range []struct {
		show    string
		videos  []specialVideo
		title   string
		episode int
		want    int // 0: not mapped, the add-on is not asked
	}{
		{"Haikyu!!", jinxHaikyuIMDb, "Lev Appears!", 1, 1},
		{"Haikyu!!", jinxHaikyuIMDb, "VS Failing Marks", 2, 3},
		{"Haikyu!!", jinxHaikyuIMDb, "Special Feature! The Spring Tournament of Their Youth", 3, 5},
		{"Haikyu!!", jinxHaikyuIMDb, "Land vs. Air", 4, 8},
		{"Haikyu!!", jinxHaikyuIMDb, "The Path of the Ball", 5, 9},
		{"Haikyu!!", jinxHaikyuIMDb, "Haikyu Quest Picture Drama - Episode 1: Before the Boss Battle", 6, 11},
		{"Haikyu!!", jinxHaikyuIMDb, "Haikyu Quest Picture Drama - Episode 3: Magic of Resurrection", 8, 13},
		{"Kuroko's Basketball", jinxKurokoIMDb, "NG-Shuu 1", 1, 1},
		{"Kuroko's Basketball", jinxKurokoIMDb, "NG-Shuu 2", 2, 2},
		{"Kuroko's Basketball", jinxKurokoIMDb, "NG-Shuu 2 1", 13, 13},
		{"Kuroko's Basketball", jinxKurokoIMDb, "Would You Like to Chat?", 23, 23},
		{"Kuroko's Basketball", jinxKurokoIMDb, "The Best Present(Episode 75.5)", 33, 40},
		{"Kuroko's Basketball", jinxKurokoIMDb, "Last Game NG Shuu", 34, 38},
		{"Kuroko's Basketball", jinxKurokoIMDb, "Let's Talk", 35, 0},
	} {
		req := core.SearchRequest{Type: core.TypeEpisode, Title: c.show, Season: 0, Episode: c.episode, EpisodeTitle: c.title}
		got, ok := mapSpecial(req, c.videos)
		if !ok {
			got = 0
		}
		if got != c.want {
			t.Errorf("mapSpecial(%q, %d) = %d, want %d", c.title, c.episode, got, c.want)
		}
	}
}
