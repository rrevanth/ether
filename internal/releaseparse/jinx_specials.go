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
	// reBareDashNumber is an anime-style bare episode number: "Show - 01 - 1080p".
	reBareDashNumber = regexp.MustCompile(`(?:^|\s)-\s\d{1,4}(?:\s?v\d{1,2})?(?:\s-\s|\s*[\[(]|\s*$)`)
	reTitleStop      = regexp.MustCompile(`(?i)[\[(]|\b(\d{3,4}p|4k|uhd|blu-?ray|bdrip|bd|brrip|web|webrip|web-dl|hdtv|dvd|dvdrip|remux|x26[45]|h\.?26[45]|hevc|avc|av1|aac|ac3|eac3|ddp?|dts|flac|opus|multi|dual|custom|nf|cr|amzn|dsnp|hmax|10bits?|8bit|hdr|sdr|dv|dovi|proper|repack|ita|eng|jpn|jap|vostfr|sub|subs|dub|mkv|mp4)\b`)
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

// RegularNumbered reports a file or release name numbered like a regular
// episode with no season ("One Piece - 590 (NF) [1080p]", "One Piece 01"):
// a bare number, which never by itself names a special.
func RegularNumbered(name string) bool {
	base := name
	if i := strings.LastIndexAny(base, "/\\"); i >= 0 {
		base = base[i+1:]
	}
	if i := strings.LastIndex(base, "."); i > 0 && len(base)-i <= 5 {
		base = base[:i] // extension
	}
	info := Parse(base)
	if len(info.Seasons) > 0 {
		return false
	}
	return len(info.Episodes) > 0 || len(info.AbsoluteEpisodes) > 0 || reBareDashNumber.MatchString(base)
}

// NamesTitle reports a name that spells out at least half of want's words
// (the show's names and filler words aside): positive evidence that it is
// that episode, where EpisodeTitleMatches only finds no contradiction.
func NamesTitle(name, want string, shows ...string) bool {
	wanted := specialWords(want)
	have := specialWords(strings.NewReplacer(".", " ", "_", " ").Replace(name))
	for _, show := range shows {
		for w := range specialWords(show) {
			delete(wanted, w)
		}
	}
	if len(wanted) == 0 {
		return false
	}
	shared := 0
	for w := range wanted {
		if have[w] {
			shared++
		}
	}
	return shared*2 >= len(wanted)
}

// Specials named by another special or by a bare group count (jinx). Release
// groups number specials their own way ("OVA - 2 (s00e02)" is TMDB's S00E13
// "The Sudden Visitor"), so a file's number alone can name the wrong
// special. Two tells, both judged on the file name and its folder:
//   - the name spells out another special of the show (at least half of
//     that title's words) better than the wanted one;
//   - it names no title at all, only a group's count ("OAD 02", "22.5",
//     "S01SP01"), with no S00Exx number to go by.
// A title in another language (romaji, Italian) names neither, and stays.

var (
	reBracketed   = regexp.MustCompile(`\[[^\]]*\]|\{[^}]*\}`)
	reGroupSuffix = regexp.MustCompile(`-[A-Za-z0-9]{2,15}$`)
	reExtension   = regexp.MustCompile(`\.[A-Za-z0-9]{2,4}$`)
	reNoiseWord   = regexp.MustCompile(`(?i)^(s\d{1,3}(e\d{1,4})?|e\d{1,4}|s\d{1,3}sp\d{1,3}|sp\d{1,3}|ep\d{1,4}|\d+x\d+|\d{3,4}[pi]|x26[45]|h26[45]|\d+bits?|v\d{1,2}|(aac|ddp?|ac3|eac3|dts|opus|flac)\d*|[0-9a-f]{8}|\d+)$`)
)

var specialNoise = map[string]bool{}

func init() {
	for _, w := range strings.Fields(`ova ovas oad oads ona onas specials season seasons series episodes extra extras bonus complete
		collection batch pack movies movie raws raw bd bluray brrip bdrip bdremux remux web webrip webdl dl hdtv dvd dvdrip hevc avc av1
		aac ac3 eac3 dts flac opus truehd atmos multi dual audio sub subs subbed dubbed dub eng english jpn jap japanese ita vostfr custom nf cr
		amzn dsnp hmax hdr sdr dv dovi proper repack uhd 4k mkv mp4 avi end final recap mixed hybrid fansub sd hd fhd no wa wo ni`) {
		specialNoise[w] = true
	}
}

func similarWord(a, b string) bool {
	if a == b {
		return true
	}
	if len(a) > len(b) {
		a, b = b, a
	}
	return len(a) >= 4 && strings.HasPrefix(b, a) // "barto" and "bartos"
}

func dropShowWords(words map[string]bool, shows []string) {
	for _, show := range shows {
		for s := range specialWords(show) {
			for w := range words {
				if similarWord(w, s) {
					delete(words, w)
				}
			}
		}
	}
}

func meaningful(words map[string]bool) map[string]bool {
	for w := range words {
		if len(w) < 2 || specialNoise[w] || reNoiseWord.MatchString(w) {
			delete(words, w)
		}
	}
	return words
}

// nameTitleWords are the title words a file's name and its folder's name
// carry: no tags, groups, numbers, technical words or the show's names.
func nameTitleWords(filePath string, shows []string) map[string]bool {
	segs := strings.Split(strings.ReplaceAll(filePath, "\\", "/"), "/")
	parts := []string{reExtension.ReplaceAllString(segs[len(segs)-1], "")}
	if len(segs) >= 2 {
		parts = append(parts, segs[len(segs)-2])
	}
	out := map[string]bool{}
	for _, p := range parts {
		p = reGroupSuffix.ReplaceAllString(strings.TrimSpace(reBracketed.ReplaceAllString(p, " ")), "")
		for w := range specialWords(strings.NewReplacer(".", " ", "_", " ").Replace(p)) {
			out[w] = true
		}
	}
	dropShowWords(out, shows)
	return meaningful(out)
}

// agreement is the share of title's words that words carry; -1 when title
// has none of its own.
func agreement(words map[string]bool, title string, shows []string) float64 {
	tw := specialWords(title)
	dropShowWords(tw, shows)
	tw = meaningful(tw)
	if len(tw) == 0 {
		return -1
	}
	shared := 0
	for t := range tw {
		for w := range words {
			if similarWord(w, t) {
				shared++
				break
			}
		}
	}
	return float64(shared) / float64(len(tw))
}

// SpecialNamesAnother reports a file whose name or folder spells out
// another of the show's specials (others) better than the wanted one.
func SpecialNamesAnother(filePath, want string, others, shows []string) bool {
	words := nameTitleWords(filePath, shows)
	if len(words) == 0 || strings.TrimSpace(want) == "" {
		return false
	}
	w := agreement(words, want, shows)
	for _, o := range others {
		if strings.TrimSpace(o) == "" || strings.EqualFold(strings.TrimSpace(o), strings.TrimSpace(want)) {
			continue
		}
		if a := agreement(words, o, shows); a >= 0.5 && a > w {
			return true
		}
	}
	return false
}

// SpecialCountOnly reports a file that names no title, only a group's count
// ("[Kawaiika-Raws] Shingeki no Kyojin OAD 02", "Kuroko no Basket - 22.5"),
// and no S00Exx number.
func SpecialCountOnly(filePath string, shows []string) bool {
	segs := strings.Split(strings.ReplaceAll(filePath, "\\", "/"), "/")
	base := strings.NewReplacer(".", " ", "_", " ").Replace(segs[len(segs)-1])
	if reSpecialMarker.MatchString(base) {
		return false
	}
	return len(nameTitleWords(filePath, shows)) == 0
}
