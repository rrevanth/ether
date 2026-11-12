package releaseparse

import (
	"regexp"
	"strings"
	"unicode"
)

// Specials (season 0) are numbered differently by TMDB, IMDb and every
// release group, and their titles are written in many lengths ("Episode of
// Luffy" for "Episode of Luffy: Adventure on Hand Island"). A title named
// after a release's S00Exx agrees with TMDB's when at least half of its
// words are TMDB's, the show's own name aside; a release naming no title
// agrees. Used by the resolver, the debrid file check and playback (jinx).

var (
	reSpecialMarker = regexp.MustCompile(`(?i)\bS0{1,2}E\d{1,4}\b`)
	reSpecialsWord  = regexp.MustCompile(`(?i)\b(specials?|ovas?|oads?|extras)\b`)
	reTitleStop     = regexp.MustCompile(`(?i)[\[(]|\b(\d{3,4}p|4k|uhd|blu-?ray|bdrip|bd|brrip|web|webrip|web-dl|hdtv|dvd|dvdrip|remux|x26[45]|h\.?26[45]|hevc|avc|av1|aac|ac3|eac3|ddp?|dts|flac|opus|multi|dual|custom|nf|cr|amzn|dsnp|hmax|10bits?|8bit|hdr|sdr|dv|dovi|proper|repack|ita|eng|jpn|jap|vostfr|sub|subs|dub|mkv|mp4)\b`)
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

// SpecialReleaseTitle is the title a name gives after its S00Exx, or "".
func SpecialReleaseTitle(name string) string {
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

// NamesSpecial reports a name marked as a special: S00Exx or a
// Specials/OVA/OAD word.
func NamesSpecial(name string) bool {
	return reSpecialsWord.MatchString(name) || reSpecialMarker.MatchString(name)
}

// SpecialTitleAgrees reports whether the title name gives after its S00Exx
// fits want; shows are the show's names, which count for neither side.
func SpecialTitleAgrees(name, want string, shows ...string) bool {
	title := specialWords(SpecialReleaseTitle(name))
	wanted := specialWords(want)
	for _, show := range shows {
		for w := range specialWords(show) {
			delete(title, w)
			delete(wanted, w)
		}
	}
	if len(title) == 0 || len(wanted) == 0 {
		return true
	}
	shared := 0
	for w := range title {
		if wanted[w] {
			shared++
		}
	}
	return shared*2 >= len(title)
}
