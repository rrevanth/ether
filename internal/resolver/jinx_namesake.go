package resolver

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

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
//     Even the owner turns away a release carrying the other show's identity:
//     its premiere year ("One Piece S01E01 2023") or its title for this very
//     episode ("ONE.PIECE.S01E01.ROMANCE.DAWN" is the 2023 live action).
// The lookups mirror series_year.go: once per show and title, cached, before
// ranking; an unanswered lookup keeps the release (upstream's behaviour).

const maxNamesakeLookups = 4

// Namesake is another catalogue show titled like the request's show.
type Namesake struct {
	Key  core.RatingKey
	Year int
}

type namesakeVerdict struct {
	other  bool       // another show's own title is this one
	owns   bool       // of the shows titled this, the request's show is the best known
	others []Namesake // those other shows
	at     time.Time
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
	if idMatch, _ := releaseparse.CompareTitleIDs(rel.ReportedTitleIDs(), sreq.IDs); idMatch {
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
		others, owns, err := s.titleNamesake(lctx, t.ShowKey, title)
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
		s.namesakes.seen[key] = namesakeVerdict{other: len(others) > 0, owns: owns, others: others, at: now}
		s.namesakes.mu.Unlock()
	}
	s.prepareNamesakeEpisodes(ctx, t, sreq)
}

// prepareNamesakeEpisodes looks up the shared-title shows' own titles for
// the requested season and episode, once each (cached for the process).
func (s *Service) prepareNamesakeEpisodes(ctx context.Context, t *core.Title, sreq core.SearchRequest) {
	if s.namesakeEpisodeTitle == nil || sreq.Episode <= 0 {
		return
	}
	s.namesakes.mu.Lock()
	v := s.namesakes.seen[namesakeKey(t.ShowKey, sreq.Title)]
	s.namesakes.mu.Unlock()
	for i, ns := range v.others {
		if i == maxNamesakeLookups {
			break
		}
		k := namesakeEpisodeKey(ns.Key, sreq.Season, sreq.Episode)
		if _, ok := s.namesakeEpisodes.Load(k); ok {
			continue
		}
		lctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		title, err := s.namesakeEpisodeTitle(lctx, ns.Key, sreq.Season, sreq.Episode)
		cancel()
		if err == nil {
			s.namesakeEpisodes.Store(k, title)
		}
	}
}

func namesakeEpisodeKey(show core.RatingKey, season, episode int) string {
	return strconv.FormatInt(int64(show), 10) + "/" + strconv.Itoa(season) + "/" + strconv.Itoa(episode)
}

var reAnyYear = regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)

// namesTitle reports a release name that spells out title (an episode
// title), as words, not merely without contradicting it.
func namesTitle(name, title string) bool {
	norm := func(s string) string {
		s = strings.NewReplacer("'", "", "’", "", "`", "").Replace(strings.ToLower(s)) // "I'm" is written "Im"
		return " " + strings.Join(strings.FieldsFunc(s, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }), " ") + " "
	}
	t := strings.TrimSpace(norm(title))
	if len(t) < 5 || strings.HasPrefix(t, "episode ") {
		return false
	}
	return strings.Contains(norm(name), " "+t+" ")
}

// namesakeTell reports a release named by a shared title that carries the
// other show's identity: that show's premiere year (and not the request's),
// or that show's title for this very episode (and not the request's).
func (s *Service) namesakeTell(rel core.Release, t *core.Title, sreq core.SearchRequest, others []Namesake) string {
	for _, m := range reAnyYear.FindAllString(rel.Name, -1) {
		y, _ := strconv.Atoi(m)
		if !sreq.Released.IsZero() && y-sreq.Released.Year() <= 1 && sreq.Released.Year()-y <= 1 {
			continue // the episode's own air year ("One Piece - 1080 (2023)")
		}
		if sreq.Year > 0 && (y-sreq.Year > 1 || sreq.Year-y > 1) {
			for _, ns := range others {
				if ns.Year == y {
					return "names another show's year"
				}
			}
		}
	}
	for _, ns := range others {
		v, ok := s.namesakeEpisodes.Load(namesakeEpisodeKey(ns.Key, sreq.Season, sreq.Episode))
		if !ok {
			continue
		}
		if other := v.(string); other != "" && namesTitle(rel.Name, other) && !namesTitle(rel.Name, sreq.EpisodeTitle) {
			return "names another show's episode"
		}
	}
	return ""
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
		if v, ok := verdict(sreq.Title); ok && v.other {
			if !v.owns {
				return "unyeared title belongs to a better-known show"
			}
			if reason := s.namesakeTell(rel, t, sreq, v.others); reason != "" {
				return reason
			}
		}
	case altTitleOnly(rel, t, sreq):
		if v, ok := verdict(rel.Info.Title); ok && v.other {
			return "title names another show"
		}
	}
	return ""
}

// namesakeAmbiguous reports a release that names its show only by a title it
// shares with a namesake that has this same episode (One Piece S01E01 is
// both the 1999 anime's first episode and the 2023 live action's), with
// nothing in the name pointing at the request's show: not its episode title,
// not an absolute number ("- 001", "EP0001"), not its series or air year,
// not a fansub group tag. Such a release ranks after every release that
// names this show and plays only when nothing else does.
func (s *Service) namesakeAmbiguous(rel core.Release, t *core.Title, sreq core.SearchRequest) bool {
	if t == nil || !mainTitleOnly(rel, sreq) || sreq.Episode <= 0 {
		return false
	}
	s.namesakes.mu.Lock()
	v, ok := s.namesakes.seen[namesakeKey(t.ShowKey, sreq.Title)]
	s.namesakes.mu.Unlock()
	if !ok || !v.other || s.now().Sub(v.at) >= seriesYearTTL {
		return false
	}
	shared := false
	for _, ns := range v.others {
		if e, ok := s.namesakeEpisodes.Load(namesakeEpisodeKey(ns.Key, sreq.Season, sreq.Episode)); ok && e.(string) != "" {
			shared = true
			break
		}
	}
	if !shared {
		return false
	}
	switch {
	case namesTitle(rel.Name, sreq.EpisodeTitle):
		return false
	case len(rel.Info.AbsoluteEpisodes) > 0:
		return false
	case strings.HasPrefix(strings.TrimSpace(rel.Name), "["):
		return false
	}
	for _, m := range reAnyYear.FindAllString(rel.Name, -1) {
		y, _ := strconv.Atoi(m)
		if y == sreq.Year || (!sreq.Released.IsZero() && y == sreq.Released.Year()) {
			return false
		}
	}
	return true
}

// withoutAmbiguous drops releases that might be a namesake's from the Play
// Version picker when any release names the show itself: a viewer picking
// "4K Dolby Vision" for anime One Piece S01E01 must not get the 2023 live
// action. With nothing else, they stay (as Auto's fallback does).
func withoutAmbiguous(cands []candidate) []candidate {
	clear := 0
	for _, c := range cands {
		if !c.ambiguous {
			clear++
		}
	}
	if clear == 0 || clear == len(cands) {
		return cands
	}
	out := make([]candidate, 0, clear)
	for _, c := range cands {
		if !c.ambiguous {
			out = append(out, c)
		}
	}
	return out
}

// withoutUnwantedAudio drops releases whose declared audio is not an
// accepted language (a Croatian-only dub for an English title) from the Play
// Version picker when releases in an accepted language exist; Auto already
// ranks them last (langOK). With nothing else, they stay.
func withoutUnwantedAudio(cands []candidate) []candidate {
	ok := 0
	for _, c := range cands {
		if c.langOK {
			ok++
		}
	}
	if ok == 0 || ok == len(cands) {
		return cands
	}
	out := make([]candidate, 0, ok)
	for _, c := range cands {
		if c.langOK {
			out = append(out, c)
		}
	}
	return out
}
