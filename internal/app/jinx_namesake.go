package app

import (
	"context"
	"errors"

	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/releaseparse"
	"github.com/braedonsaunders/ether/internal/resolver"
)

// titleNamesake reports whether the catalogue holds a show other than show
// whose own title (or original title) is title, and whether show owns that
// title (resolver/jinx_namesake.go). Another show takes it only when it is
// clearly better known: at least twice show's TMDB votes, and ten more.
// Without vote counts, or with show missing from the results, show owns it:
// an unknown answer must not cost a show its releases.
func (a *App) titleNamesake(ctx context.Context, show core.RatingKey, title string) (others []resolver.Namesake, owns bool, err error) {
	if a.catalog == nil {
		return nil, false, errors.New("catalogue unavailable")
	}
	shows, err := a.catalog.SearchAll(ctx, title, []core.MediaType{core.TypeShow}, 20)
	if err != nil {
		return nil, false, err
	}
	var self *core.Title
	topOther := -1
	for i := range shows {
		t := &shows[i]
		if t.Key == show {
			self = t
			continue
		}
		if releaseparse.TitlesMatch(t.Title, title) || (t.OriginalTitle != "" && releaseparse.TitlesMatch(t.OriginalTitle, title)) {
			others = append(others, resolver.Namesake{Key: t.Key, Year: t.Year})
			topOther = max(topOther, t.VoteCount)
		}
	}
	owns = len(others) == 0 || self == nil || topOther <= 0 || topOther < 2*self.VoteCount+10
	return others, owns, nil
}

// namesakeEpisodeTitle is another show's title for one episode (catalogue,
// fetched from TMDB when not yet stored); "" when it has no such episode.
func (a *App) namesakeEpisodeTitle(ctx context.Context, show core.RatingKey, season, episode int) (string, error) {
	if a.catalog == nil {
		return "", errors.New("catalogue unavailable")
	}
	t, err := a.catalog.Get(ctx, core.EpisodeKey(show.TMDB(), season, episode))
	if err != nil || t == nil {
		return "", nil
	}
	return t.Title, nil
}
