package resolver

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/releaseparse"
)

// A yearless episode release with no reported ids names its show only by
// title, and titles are shared. Two rules keep it with the show the name
// means:
//   - matched only through an alternative title that is a different show's
//     own title, it is that show's: TMDB lists plain "RAMAYAN" for the 2026
//     "RAMAYAN AN ATHARV SRIVASTAVA SERIES", and "Ramayan.S01E01-44" is the
//     1987 Ramayan;
//   - matched by a main title several shows share ("Ramayan" 1987 and 2008),
//     it belongs to the one most TMDB voters know, as Sonarr maps an
//     unyeared title; the others need a series year or ids in the release.
// The lookups mirror series_year.go: once per show and title, cached, before
// ranking; an unanswered lookup keeps the release (upstream's behaviour).

const maxNamesakeLookups = 4

type namesakeVerdict struct {
	other bool // another show's own title is this one
	owns  bool // of the shows titled this, the request's show is the best known
	at    time.Time
}

type namesakes struct {
	mu   sync.Mutex
	seen map[string]namesakeVerdict
}

func namesakeKey(show core.RatingKey, title string) string {
	return strconv.FormatInt(int64(show), 10) + "/" + strings.ToLower(strings.TrimSpace(title))
}

// unvouched reports an episode release with no series year in its name and
// no reported ids: only its title says which show it is.
func unvouched(rel core.Release, sreq core.SearchRequest) bool {
	if sreq.Type != core.TypeEpisode || rel.Info.Title == "" {
		return false
	}
	if idMatch, _ := releaseparse.CompareTitleIDs(rel.TitleIDs, sreq.IDs); idMatch {
		return false
	}
	return releaseparse.EpisodeSeriesYear(rel.Name, rel.Info.Year) == 0
}

// mainTitleOnly reports an unvouched release named by the show's own title.
func mainTitleOnly(rel core.Release, sreq core.SearchRequest) bool {
	return unvouched(rel, sreq) && releaseparse.TitleMatches(rel.Info, sreq.Title)
}

// altTitleOnly reports an unvouched release whose title matches the request
// only through one of the show's alternative titles.
func altTitleOnly(rel core.Release, t *core.Title, sreq core.SearchRequest) bool {
	if !unvouched(rel, sreq) || releaseparse.TitleMatches(rel.Info, sreq.Title) {
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
		switch {
		case mainTitleOnly(rel, sreq):
			titles[sreq.Title] = true
		case altTitleOnly(rel, t, sreq):
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
		other, owns, err := s.titleNamesake(lctx, t.ShowKey, title)
		cancel()
		if err != nil {
			s.log.Debug("namesake lookup failed; keeping alternative-title matches", "show", t.ShowKey, "title", title, "err", err)
			continue
		}
		s.namesakes.mu.Lock()
		if s.namesakes.seen == nil {
			s.namesakes.seen = map[string]namesakeVerdict{}
		}
		if len(s.namesakes.seen) > 4096 {
			for k, old := range s.namesakes.seen {
				if now.Sub(old.at) >= seriesYearTTL {
					delete(s.namesakes.seen, k)
				}
			}
		}
		s.namesakes.seen[key] = namesakeVerdict{other: other, owns: owns, at: now}
		s.namesakes.mu.Unlock()
	}
}

// rejectNamesake is the ranking-time verdict of prepareNamesakes.
func (s *Service) rejectNamesake(rel core.Release, t *core.Title, sreq core.SearchRequest) string {
	if t == nil {
		return ""
	}
	verdict := func(title string) (namesakeVerdict, bool) {
		s.namesakes.mu.Lock()
		v, ok := s.namesakes.seen[namesakeKey(t.ShowKey, title)]
		s.namesakes.mu.Unlock()
		return v, ok && s.now().Sub(v.at) < seriesYearTTL
	}
	switch {
	case mainTitleOnly(rel, sreq):
		if v, ok := verdict(sreq.Title); ok && v.other && !v.owns {
			return "unyeared title belongs to a better-known show"
		}
	case altTitleOnly(rel, t, sreq):
		if v, ok := verdict(rel.Info.Title); ok && v.other {
			return "title names another show"
		}
	}
	return ""
}
