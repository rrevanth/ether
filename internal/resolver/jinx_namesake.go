package resolver

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/releaseparse"
)

// An alternative title can be another show's main title: TMDB lists plain
// "RAMAYAN" for the 2026 "RAMAYAN AN ATHARV SRIVASTAVA SERIES", and
// "Ramayan.S01E01-44" is the 1987 Ramayan. A yearless episode release with
// no reported ids that matches only through an alternative title is that
// other show's when the catalogue holds a different show whose own title it
// is. The lookup mirrors series_year.go: once per show and title, cached,
// before ranking; an unanswered lookup keeps the release (upstream's rule).

const maxNamesakeLookups = 4

func namesakeKey(show core.RatingKey, title string) string {
	return strconv.FormatInt(int64(show), 10) + "/" + strings.ToLower(strings.TrimSpace(title))
}

// altTitleOnly reports an episode release whose title matches the request
// only through one of the show's alternative titles, with no series year in
// its name and no reported ids to vouch for it.
func altTitleOnly(rel core.Release, t *core.Title, sreq core.SearchRequest) bool {
	if sreq.Type != core.TypeEpisode || rel.Info.Title == "" {
		return false
	}
	if idMatch, _ := releaseparse.CompareTitleIDs(rel.TitleIDs, sreq.IDs); idMatch {
		return false
	}
	if releaseparse.EpisodeSeriesYear(rel.Name, rel.Info.Year) != 0 {
		return false
	}
	if releaseparse.TitleMatches(rel.Info, sreq.Title) {
		return false
	}
	if t != nil && t.ShowOriginalTitle != "" && releaseparse.TitleMatches(rel.Info, t.ShowOriginalTitle) {
		return false
	}
	alts := sreq.AlternativeTitles
	if t != nil {
		alts = append(append([]string(nil), alts...), t.ShowAlternativeTitles...)
	}
	return matchesAlternativeTitle(rel.Info, alts)
}

// prepareNamesakes asks the catalogue, before ranking, whether each title
// that matched only as an alternative names another show.
func (s *Service) prepareNamesakes(ctx context.Context, t *core.Title, sreq core.SearchRequest, results []core.Release) {
	if s.titleNamesake == nil || t == nil || t.ShowKey == 0 || sreq.Type != core.TypeEpisode {
		return
	}
	titles := map[string]bool{}
	for _, rel := range results {
		if altTitleOnly(rel, t, sreq) {
			titles[rel.Info.Title] = true
		}
	}
	now := s.now()
	lookups := 0
	for title := range titles {
		key := namesakeKey(t.ShowKey, title)
		s.namesakes.mu.Lock()
		v, ok := s.namesakes.seen[key]
		s.namesakes.mu.Unlock()
		if ok && now.Sub(v.at) < seriesYearTTL {
			continue
		}
		if lookups == maxNamesakeLookups {
			break
		}
		lookups++
		lctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		conflict, err := s.titleNamesake(lctx, t.ShowKey, title)
		cancel()
		if err != nil {
			s.log.Debug("namesake lookup failed; keeping alternative-title matches", "show", t.ShowKey, "title", title, "err", err)
			continue
		}
		s.namesakes.mu.Lock()
		if s.namesakes.seen == nil {
			s.namesakes.seen = map[string]seriesYearVerdict{}
		}
		if len(s.namesakes.seen) > 4096 {
			for k, old := range s.namesakes.seen {
				if now.Sub(old.at) >= seriesYearTTL {
					delete(s.namesakes.seen, k)
				}
			}
		}
		s.namesakes.seen[key] = seriesYearVerdict{conflict: conflict, at: now}
		s.namesakes.mu.Unlock()
	}
}

// rejectNamesake is the ranking-time verdict of prepareNamesakes.
func (s *Service) rejectNamesake(rel core.Release, t *core.Title, sreq core.SearchRequest) string {
	if t == nil || !altTitleOnly(rel, t, sreq) {
		return ""
	}
	s.namesakes.mu.Lock()
	v, ok := s.namesakes.seen[namesakeKey(t.ShowKey, rel.Info.Title)]
	s.namesakes.mu.Unlock()
	if ok && v.conflict && s.now().Sub(v.at) < seriesYearTTL {
		return "title names another show"
	}
	return ""
}
