package stremio

import (
	"strings"
	"unicode"

	"github.com/braedonsaunders/ether/internal/core"
)

// TMDB and IMDb title the same special differently ("Land vs. Air" against
// a listing full of "Haikyu!! ..." entries, "Lev Appears!" against "The
// Arrival of Lev"). matchSpecialTitle wants two shared words and half of
// both titles; when it finds nothing, jinxMatchSpecial weighs each shared
// word by how rare it is in the series' listing: a word only one IMDb
// special uses counts 1, a word all of them use counts little. The best
// entry must reach jinxSpecialsMinScore and beat every other (jinx).
const jinxSpecialsMinScore = 1.0

var jinxSpecialStopWords = map[string]bool{"the": true, "a": true, "an": true, "of": true, "and": true, "to": true, "in": true, "on": true,
	"vs": true, "ova": true, "oad": true, "ona": true, "special": true, "episode": true, "ep": true,
	"lets": true, "its": true, "we": true, "you": true, "us": true, "shall": true, "would": true, "like": true, "do": true, "can": true}

// mapSpecial is upstream's strict title match, then the weighted fallback.
func mapSpecial(req core.SearchRequest, videos []specialVideo) (int, bool) {
	if n, ok := matchSpecialTitle(req, videos); ok {
		return n, true
	}
	return jinxMatchSpecial(req.EpisodeTitle, req.Episode, videos)
}

func jinxSpecialWords(s string) []string {
	return strings.FieldsFunc(strings.ToLower(strings.ReplaceAll(s, "'", "")), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

// jinxSpecialKey is a title's exact-match form: every word, without an OVA/OAD/ONA label.
func jinxSpecialKey(s string) string {
	var out []string
	for _, w := range jinxSpecialWords(s) {
		if w != "ova" && w != "oad" && w != "ona" {
			out = append(out, strings.TrimLeft(w, "0"))
		}
	}
	return strings.Join(out, " ")
}

func jinxSpecialTokens(s string) []string {
	var out []string
	for _, w := range jinxSpecialWords(s) {
		if jinxSpecialStopWords[w] {
			continue
		}
		if t := strings.TrimLeft(w, "0"); t != "" && strings.Trim(w, "0123456789") == "" {
			w = t // "01" and "1" are the same number
		}
		out = append(out, w)
	}
	return out
}

func jinxMatchSpecial(title string, episode int, videos []specialVideo) (int, bool) {
	want := jinxSpecialTokens(title)
	key := jinxSpecialKey(title)
	toks := make([]map[string]bool, len(videos))
	df := map[string]int{}
	for i, v := range videos {
		toks[i] = map[string]bool{}
		for _, w := range jinxSpecialTokens(v.Name) {
			toks[i][w] = true
		}
		for w := range toks[i] {
			df[w]++
		}
	}
	if key != "" {
		same := 0
		for _, v := range videos {
			if jinxSpecialKey(v.Name) == key {
				if same != 0 && same != v.Episode {
					return 0, false
				}
				same = v.Episode
			}
		}
		if same > 0 {
			return same, true
		}
	}
	if len(want) == 0 {
		// "Episode 3" against "Episode 3": no words, the same number.
		for i, v := range videos {
			if v.Episode == episode && len(toks[i]) == 0 {
				return episode, true
			}
		}
		return 0, false
	}
	wantSet := map[string]bool{}
	for _, w := range want {
		wantSet[w] = true
	}
	best, bestEp, second := 0.0, 0, 0.0
	for i, v := range videos {
		score := 0.0
		for w := range toks[i] {
			if wantSet[w] {
				score += 1 / float64(df[w])
			}
		}
		if score > best {
			best, bestEp, second = score, v.Episode, best
		} else if score > second {
			second = score
		}
	}
	if best < jinxSpecialsMinScore || best-second < 1e-9 {
		return 0, false
	}
	return bestEp, true
}
