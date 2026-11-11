package stremio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/braedonsaunders/ether/internal/core"
)

// Specials (season 0) are numbered differently by TMDB, which Ether's
// catalogue follows, and IMDb, which add-ons answer by: Haikyu!!'s TMDB
// S00E05 "The Path of the Ball" is IMDb S00E09. Asking an add-on for
// tt…:0:5 would play another special. Before asking, the TMDB episode's
// title is matched against the IMDb specials Cinemeta lists; without a
// confident match the add-on is not asked at all.

const (
	specialsTTL      = 24 * time.Hour
	specialsMissTTL  = time.Hour
	specialsMinScore = 1.0 // one shared word no other IMDb special uses
)

var specialStopWords = map[string]bool{"the": true, "a": true, "an": true, "of": true, "and": true, "to": true, "in": true, "on": true,
	"vs": true, "ova": true, "oad": true, "ona": true, "special": true, "episode": true, "ep": true,
	"lets": true, "its": true, "we": true, "you": true, "us": true, "shall": true, "would": true, "like": true, "do": true, "can": true}

type specialVideo struct {
	Episode int
	Name    string
}

type specialsEntry struct {
	at     time.Time
	ttl    time.Duration
	videos []specialVideo
}

// imdbSpecials caches Cinemeta's season-0 listing per IMDb series id.
type imdbSpecials struct {
	base   string
	client *http.Client
	now    func() time.Time

	mu   sync.Mutex
	byID map[string]specialsEntry
}

var cinemetaSpecials = &imdbSpecials{base: "https://v3-cinemeta.strem.io", client: &http.Client{Timeout: 5 * time.Second}, now: time.Now, byID: map[string]specialsEntry{}}

func (c *imdbSpecials) videos(ctx context.Context, imdb string) []specialVideo {
	now := c.now()
	c.mu.Lock()
	e, ok := c.byID[imdb]
	c.mu.Unlock()
	if ok && now.Sub(e.at) < e.ttl {
		return e.videos
	}
	vids, err := c.fetch(ctx, imdb)
	e = specialsEntry{at: now, ttl: specialsTTL, videos: vids}
	if err != nil {
		e.ttl = specialsMissTTL
	}
	c.mu.Lock()
	c.byID[imdb] = e
	c.mu.Unlock()
	return vids
}

func (c *imdbSpecials) fetch(ctx context.Context, imdb string) ([]specialVideo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/meta/series/"+imdb+".json", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cinemeta: %s", resp.Status)
	}
	var meta struct {
		Meta struct {
			Videos []struct {
				Season  int    `json:"season"`
				Episode int    `json:"episode"`
				Name    string `json:"name"`
				Title   string `json:"title"`
			} `json:"videos"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(&meta); err != nil {
		return nil, err
	}
	var out []specialVideo
	for _, v := range meta.Meta.Videos {
		if v.Season == 0 && v.Episode > 0 {
			out = append(out, specialVideo{Episode: v.Episode, Name: cmpOr(v.Name, v.Title)})
		}
	}
	return out, nil
}

func cmpOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// specialKey is a title's exact-match form: every word, lower-cased, without
// punctuation or an OVA/OAD/ONA label.
func specialKey(s string) string {
	var out []string
	for _, w := range strings.FieldsFunc(strings.ToLower(strings.ReplaceAll(s, "'", "")), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		if w != "ova" && w != "oad" && w != "ona" {
			out = append(out, strings.TrimLeft(w, "0"))
		}
	}
	return strings.Join(out, " ")
}

func specialTokens(s string) []string {
	var out []string
	for _, w := range strings.FieldsFunc(strings.ToLower(strings.ReplaceAll(s, "'", "")), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		if specialStopWords[w] {
			continue
		}
		if t := strings.TrimLeft(w, "0"); t != "" && strings.Trim(w, "0123456789") == "" {
			w = t // "01" and "1" are the same number
		}
		out = append(out, w)
	}
	return out
}

// matchSpecial maps a TMDB special (title, number) to the IMDb episode
// number: the listing's entry with the same words, else the entry sharing
// the most distinctive words. Each shared word counts 1/n, n being how
// many IMDb specials use it, so "Lev" (one entry) outweighs "Haikyu" (all
// of them). The best entry must reach specialsMinScore and beat every
// other. A TMDB title with no usable words maps only when the IMDb entry
// at the same number has none either (both "Episode 3").
func matchSpecial(title string, episode int, videos []specialVideo) (int, bool) {
	want := specialTokens(title)
	key := specialKey(title)
	toks := make([]map[string]bool, len(videos))
	df := map[string]int{}
	for i, v := range videos {
		toks[i] = map[string]bool{}
		for _, w := range specialTokens(v.Name) {
			toks[i][w] = true
		}
		for w := range toks[i] {
			df[w]++
		}
		if key != "" && specialKey(v.Name) == key {
			return v.Episode, true
		}
	}
	if len(want) == 0 {
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
	if best < specialsMinScore || best-second < 1e-9 {
		return 0, false
	}
	return bestEp, true
}

// imdbSpecialRequest is req with the special renumbered to IMDb's order
// for asking add-ons by IMDb id; ok is false when it cannot be mapped.
// Requests other than an IMDb-identified special pass through unchanged.
func imdbSpecialRequest(ctx context.Context, req core.SearchRequest) (core.SearchRequest, bool) {
	imdb := IMDbID(req.IDs.IMDB)
	if req.Type != core.TypeEpisode || req.Season != 0 || req.Episode <= 0 || imdb == "" {
		return req, true
	}
	ep, ok := matchSpecial(req.EpisodeTitle, req.Episode, cinemetaSpecials.videos(ctx, imdb))
	if !ok {
		return req, false
	}
	req.Episode = ep
	return req, true
}
