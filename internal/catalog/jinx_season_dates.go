package catalog

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/tmdb"
)

// TMDB leaves the air date off some seasons whose episodes all have one: One
// Piece's arc seasons (Whiskey Peak, Drum Island … Egghead, Elbaph). An
// undated season is "unreleased" to every listing, so those seasons, and
// every episode in them, disappeared, and a hidden season is never opened, so
// its episodes are never fetched to date it. Upstream dates only an undated
// Specials season from its children; numbered seasons get the same here:
//   - a season's episode fetch dates the season by its first aired episode;
//   - a show refresh dates an undated season TMDB places at or before the
//     last aired episode's season by its stored episodes, else provisionally
//     by the latest earlier dated season (or the show's premiere), so it lists
//     and its first open fetches the real date.

// datedFromEpisodes gives an undated numbered season its earliest aired
// episode's date.
func datedFromEpisodes(season core.Title, eps []core.Title) core.Title {
	if !season.Released.IsZero() || season.Season == 0 {
		return season
	}
	var first time.Time
	for _, e := range eps {
		if !e.Released.IsZero() && (first.IsZero() || e.Released.Before(first)) {
			first = e.Released
		}
	}
	if !first.IsZero() {
		season.Released, season.Year = first, first.Year()
	}
	return season
}

// datedSummaries dates the undated season summaries of a show refresh (see
// above). seasons is modified in place.
func (c *Catalog) datedSummaries(ctx context.Context, show core.Title, d *tmdb.TVDetails, seasons []core.Title) {
	if d == nil || d.LastEpisodeToAir == nil {
		return
	}
	last := d.LastEpisodeToAir.SeasonNumber
	order := make([]int, len(seasons))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return seasons[order[a]].Season < seasons[order[b]].Season })
	prev := show.Released
	for _, i := range order {
		s := &seasons[i]
		if s.Season == 0 {
			continue
		}
		if !s.Released.IsZero() {
			prev = s.Released
			continue
		}
		if s.Season > last || s.ChildCount == 0 {
			continue
		}
		if first := c.firstStoredAirDate(ctx, show.Key, s.Season); !first.IsZero() {
			s.Released = first
		} else if !prev.IsZero() {
			s.Released = prev
		} else {
			continue
		}
		s.Year = s.Released.Year()
		prev = s.Released
	}
}

func (c *Catalog) firstStoredAirDate(ctx context.Context, show core.RatingKey, season int) time.Time {
	var ts sql.NullInt64
	if c.db == nil {
		return time.Time{}
	}
	if err := c.db.QueryRowContext(ctx, `SELECT MIN(released_ts) FROM catalog_titles WHERE type='episode' AND show_key=? AND season=? AND released_ts<>0`, int64(show), season).Scan(&ts); err != nil || !ts.Valid {
		return time.Time{}
	}
	return time.Unix(ts.Int64, 0).UTC()
}
