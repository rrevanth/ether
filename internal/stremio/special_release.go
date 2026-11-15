package stremio

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/releaseparse"
)

var specialNumberTag = regexp.MustCompile(`(?i)\bS0{1,3}[ ._-]*E(\d{1,4})\b`)

// specialRelease validates the add-on's own filename before translating the
// identity to the catalogue order. The original IMDb number is retained for
// debrid's independent file validation.
func specialRelease(rel core.Release, req core.SearchRequest, file, torrent string, asked int, trustAddon bool) (core.Release, bool) {
	shows := append([]string{req.Title, req.OriginalTitle}, req.AlternativeTitles...)
	actual := file
	if actual == "" {
		actual = torrent
	}
	if actual == "" {
		actual = rel.Name
	}
	info := releaseparse.Parse(actual)
	packInfo := releaseparse.Parse(torrent)
	// A regular-season pack never supplies a special; a season-zero pack
	// ("[BlackRabbit] Kuroko's Basketball (2012) - S00") is a specials pack (jinx).
	seasonZeroPack := len(packInfo.Seasons) == 1 && packInfo.Seasons[0] == 0
	if len(packInfo.Seasons) > 0 && !seasonZeroPack && !releaseparse.NamesSpecial(torrent) {
		return rel, false
	}
	if len(info.Seasons) > 0 && (!slices.Contains(info.Seasons, 0) || len(info.Seasons) != 1) {
		return rel, false
	}
	if !releaseparse.SpecialMatches(actual, req.EpisodeTitle, req.SpecialTitles, shows) {
		return rel, false
	}
	// The add-on vouches for a labelled special it answered for a mapped
	// IMDb number (romaji-titled OVAs; see below) (jinx).
	vouched := trustAddon && asked > 0 && releaseparse.NamesSpecial(actual)
	expected := req.Episode
	if asked > 0 {
		expected = asked
	}
	if len(info.Seasons) > 0 {
		if !releaseparse.EpisodeMatches(info, 0, expected) {
			return rel, false
		}
	} else {
		// Bare regular numbering never becomes season zero on the add-on's word.
		if len(info.Episodes) > 0 || len(info.AbsoluteEpisodes) > 0 {
			return rel, false
		}
		// A labelled special the add-on answered for the IMDb special it was
		// asked about is that special even when it is titled in romaji
		// ("Haikyuu!! Tokushuu! Haru-kou Volley ni Kaketa Seishun - OVA" for
		// TMDB "Special Feature! The Spring Tournament of Their Youth") (jinx).
		if !vouched && !releaseparse.SpecialTitleEvidence(actual, req.EpisodeTitle, shows...) {
			return rel, false
		}
	}
	// Use the validated special's name, retaining technical facts from the pack.
	rel.Name = actual
	rel.Name = specialNumberTag.ReplaceAllString(rel.Name, fmt.Sprintf("S00E%02d", req.Episode))
	rel.Info = describe(rel.Name, file, torrent)
	if prefix := releaseparse.SpecialSeriesTitle(actual); prefix != "" && releaseparse.TitleMatchesAny(core.ReleaseInfo{Title: prefix}, req.Title, shows) {
		rel.Info.Title = prefix
	}
	// "Haikyuu!! Tokushuu! Haru Kou Volley ni Kaketa Seishun - OVA" parses as one
	// long title; a vouched special that begins with the show's name is the show's.
	if vouched && !releaseparse.TitleMatchesAny(rel.Info, req.Title, shows) {
		name := releaseparse.NormalizeTitle(releaseparse.Parse(actual).Title)
		for _, show := range shows {
			if n := releaseparse.NormalizeTitle(show); n != "" && strings.HasPrefix(name, n+" ") {
				rel.Info.Title = show
				break
			}
		}
	}
	rel.Info.Seasons, rel.Info.Episodes, rel.Info.FullSeason = []int{0}, []int{req.Episode}, false
	rel.Info.AbsoluteEpisodes = nil
	rel.Target = &core.ReleaseTarget{Season: 0, Episode: req.Episode, EpisodeTitle: req.EpisodeTitle, ShowTitles: shows, SpecialTitles: req.SpecialTitles, AddonSpecialEpisode: asked}
	return rel, true
}

// addonAnswerAgrees reports whether the add-on's answer for a special may
// be taken on its word: every file in it that neither an S00Exx/0x number
// nor TMDB's title confirms must name the same special. Torrentio answers
// One Piece IMDb S00E42 with "Dai Tettei Kaibou! Kouzuki Oden Densetsu!"
// and another special, "Momonosuke no Mei Shougun e no Michi - SP"; neither
// can be told from TMDB's English title, so neither is trusted (jinx).
func addonAnswerAgrees(names []string, req core.SearchRequest) bool {
	shows := append([]string{req.Title, req.OriginalTitle}, req.AlternativeTitles...)
	// Titles are grouped by overlap, not equality: a release's leftover tags
	// ("10bit", "web1080p") must not split one special into two.
	var groups []map[string]bool
	for _, name := range names {
		info := releaseparse.Parse(name)
		if len(info.Seasons) > 0 || len(info.Episodes) > 0 || len(info.AbsoluteEpisodes) > 0 || releaseparse.SpecialTitleEvidence(name, req.EpisodeTitle, shows...) {
			continue
		}
		words := releaseparse.SpecialTitleWords(info.Title, shows...)
		if len(words) == 0 {
			continue
		}
		same := false
		for _, g := range groups {
			shared := 0
			for w := range words {
				if g[w] {
					shared++
				}
			}
			if shared*5 >= min(len(words), len(g))*4 {
				same = true
				break
			}
		}
		if !same {
			groups = append(groups, words)
		}
	}
	return len(groups) <= 1
}
