package resolver

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/braedonsaunders/ether/internal/core"
)

// Specials (season 0) are numbered differently by TMDB, IMDb and every
// release group: One Piece's "S00E01" is TMDB's 1999 summary special, A&C's
// "S00E02 - Movie 01" and Crunchyroll's "S00E01 One Piece Heroines". The
// number alone does not identify a special, so a release for one must
//   - say it is a special: season 0, or a Specials/OVA/OAD word (a batch
//     "Ep.001-165" of regular episodes is not), and
//   - not name another special: a title after its S00Exx must agree with
//     TMDB's (at least half of its words in TMDB's title, so "Episode of
//     Luffy" fits "Episode of Luffy: Adventure on Hand Island"); a release
//     naming no title is judged by its number alone.
// This replaces upstream's episode-title rule for season 0, which rejected
// partial titles and let multi-language releases through unchecked.

var (
	reSpecialMarker = regexp.MustCompile(`(?i)\bS0{1,2}E\d{1,4}\b`)
	reSpecialsWord  = regexp.MustCompile(`(?i)\b(specials?|ovas?|oads?|extras)\b`)
	reTitleStop     = regexp.MustCompile(`(?i)[\[(]|\b(\d{3,4}p|4k|uhd|blu-?ray|bdrip|bd|brrip|web|webrip|web-dl|hdtv|dvd|dvdrip|remux|x26[45]|h\.?26[45]|hevc|avc|av1|aac|ac3|eac3|ddp?|dts|flac|opus|multi|dual|custom|nf|cr|amzn|dsnp|hmax|10bits?|8bit|hdr|sdr|dv|dovi|proper|repack|ita|eng|jpn|jap|vostfr|sub|subs|dub)\b`)
)

var specialStop = map[string]bool{"the": true, "a": true, "an": true, "of": true, "and": true, "to": true, "in": true, "on": true,
	"at": true, "for": true, "is": true, "episode": true, "ep": true, "special": true, "sp": true, "part": true, "vs": true}

func specialWords(s string) map[string]bool {
	s = strings.NewReplacer("'", "", "’", "").Replace(strings.ToLower(s))
	out := map[string]bool{}
	for _, w := range strings.FieldsFunc(s, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		if specialStop[w] {
			continue
		}
		if n := strings.TrimLeft(w, "0"); n != "" && strings.Trim(w, "0123456789") == "" {
			w = n // "01" and "1" are the same number
		}
		out[w] = true
	}
	return out
}

// specialReleaseTitle is the title a release names after its S00Exx, or "".
func specialReleaseTitle(name string) string {
	spaced := strings.NewReplacer(".", " ", "_", " ").Replace(name)
	loc := reSpecialMarker.FindStringIndex(spaced)
	if loc == nil {
		return ""
	}
	tail := strings.TrimLeft(spaced[loc[1]:], " -–:")
	if m := reTitleStop.FindStringIndex(tail); m != nil {
		tail = tail[:m[0]]
	}
	return strings.TrimSpace(strings.Trim(tail, " -–:"))
}

func rejectSpecial(rel core.Release, sreq core.SearchRequest) string {
	if sreq.Type != core.TypeEpisode || sreq.Season != 0 || sreq.Episode <= 0 {
		return ""
	}
	marked := reSpecialsWord.MatchString(rel.Name) || reSpecialMarker.MatchString(rel.Name)
	for _, s := range rel.Info.Seasons {
		marked = marked || s == 0
	}
	if !marked {
		return "names no special"
	}
	title := specialWords(specialReleaseTitle(rel.Name))
	want := specialWords(sreq.EpisodeTitle)
	// The show's own name ("One Piece Heroines", "...Summary One Piece") is no
	// evidence either way.
	for _, show := range append([]string{sreq.Title, sreq.OriginalTitle}, sreq.AlternativeTitles...) {
		for w := range specialWords(show) {
			delete(title, w)
			delete(want, w)
		}
	}
	if len(title) == 0 || len(want) == 0 {
		return ""
	}
	shared := 0
	for w := range title {
		if want[w] {
			shared++
		}
	}
	if shared*2 < len(title) {
		return "special's title differs"
	}
	return ""
}
