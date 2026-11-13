package resolver

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/releaseparse"
)

// Season and special names tell releases apart that numbers cannot (jinx):
//   - "Bleach - Thousand-Year Blood War - S01E01" is the first episode of the
//     season TMDB names "Thousand-Year Blood War" (Bleach S02), not S01E01:
//     a release whose title adds another season's name is that season's;
//   - "Lost Girls OVA - 01 (s00e06)" is the special TMDB titles "Lost
//     Girls", not its S00E06: a special's release naming another special
//     is that one (releaseparse.SpecialNamesAnother; the debrid file check
//     applies the same test to the file actually opened).
// The names come from the catalogue once per show (and season), cached; an
// unanswered lookup leaves the request as before.

type seasonNameEntry struct {
	titles map[int]string
	at     time.Time
}

type seasonNames struct {
	mu   sync.Mutex
	seen map[string]seasonNameEntry
}

const seasonNamesTTL = 6 * time.Hour

func (n *seasonNames) get(key string, now time.Time, load func() (map[int]string, error)) map[int]string {
	n.mu.Lock()
	if e, ok := n.seen[key]; ok && now.Sub(e.at) < seasonNamesTTL {
		n.mu.Unlock()
		return e.titles
	}
	n.mu.Unlock()
	titles, err := load()
	if err != nil {
		return nil
	}
	n.mu.Lock()
	if n.seen == nil {
		n.seen = map[string]seasonNameEntry{}
	}
	n.seen[key] = seasonNameEntry{titles: titles, at: now}
	n.mu.Unlock()
	return titles
}

func (s *Service) withSeasonNames(ctx context.Context, t *core.Title, req *core.SearchRequest) {
	if s == nil || t == nil || t.ShowKey == 0 || req.Type != core.TypeEpisode {
		return
	}
	lctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	show := t.ShowKey
	if s.showSeasonTitles != nil {
		req.SeasonTitles = s.seasonNames.get(show.String()+"/seasons", s.now(), func() (map[int]string, error) {
			return s.showSeasonTitles(lctx, show)
		})
	}
	if s.seasonEpisodeTitles != nil && req.Season == 0 && req.Episode > 0 {
		titles := s.seasonNames.get(show.String()+"/s0", s.now(), func() (map[int]string, error) {
			return s.seasonEpisodeTitles(lctx, show, 0)
		})
		req.SiblingTitles = nil
		for ep, title := range titles {
			if ep != req.Episode && title != "" {
				req.SiblingTitles = append(req.SiblingTitles, title)
			}
		}
	}
}

// showNames are the show's own names: its title, original title and the
// alternative titles that are names of the show itself (romaji "Shingeki no
// Kyojin", "Kuroko no Basuke"). An alternative that only extends another of
// its names ("Attack on Titan: Lost Girls", "HAIKYU!! LAND VS. AIR") names a
// part or special of it, and its extra words must still count.
func showNames(sreq core.SearchRequest) []string {
	all := []string{sreq.Title}
	if sreq.OriginalTitle != "" {
		all = append(all, sreq.OriginalTitle)
	}
	all = append(all, sreq.AlternativeTitles...)
	norm := make([]string, len(all))
	for i, n := range all {
		norm[i] = releaseparse.NormalizeTitle(n)
	}
	var out []string
	for i, n := range all {
		if norm[i] == "" {
			continue
		}
		extends := false
		for j := range all {
			if j != i && norm[j] != "" && len(norm[j]) < len(norm[i]) && strings.HasPrefix(norm[i]+" ", norm[j]+" ") {
				extends = true
				break
			}
		}
		if !extends {
			out = append(out, n)
		}
	}
	return out
}

// specialTargetTitles is what a special's file check needs: the show's names
// and its other specials' titles.
func specialTargetTitles(sreq core.SearchRequest) ([]string, []string) {
	if sreq.Season != 0 || sreq.Episode <= 0 {
		return nil, nil
	}
	return showNames(sreq), sreq.SiblingTitles
}

var reGenericSeason = regexp.MustCompile(`(?i)^\s*(season|series|part|book|volume|vol\.?|specials?|extras?)\b[\s\d.:-]*$`)

func rejectOtherSeasonOrSpecial(rel core.Release, sreq core.SearchRequest) string {
	if sreq.Type != core.TypeEpisode || sreq.Episode <= 0 {
		return ""
	}
	if sreq.Season == 0 {
		name := rel.Name
		if rel.FileName != "" {
			name = rel.FileName
		}
		if releaseparse.SpecialNamesAnother(name, sreq.EpisodeTitle, sreq.SiblingTitles, showNames(sreq)) {
			return "names another special"
		}
		return ""
	}
	title := strings.ToLower(releaseparse.NormalizeTitle(rel.Info.Title))
	if title == "" || len(sreq.SeasonTitles) < 2 {
		return ""
	}
	own := strings.ToLower(releaseparse.NormalizeTitle(sreq.SeasonTitles[sreq.Season]))
	if ownName := sreq.SeasonTitles[sreq.Season]; ownName == "" || reGenericSeason.MatchString(ownName) || releaseparse.TitlesMatch(ownName, sreq.Title) {
		own = "" // a season named like the show ("Bleach" S1) names nothing
	}
	for n, name := range sreq.SeasonTitles {
		if n == sreq.Season || n == 0 || reGenericSeason.MatchString(name) || releaseparse.TitlesMatch(name, sreq.Title) {
			continue
		}
		other := strings.ToLower(releaseparse.NormalizeTitle(name))
		if len(strings.Fields(other)) < 2 || (own != "" && strings.Contains(title, own)) {
			continue
		}
		if strings.Contains(" "+title+" ", " "+other+" ") {
			return "names another season (" + name + ")"
		}
	}
	return ""
}
