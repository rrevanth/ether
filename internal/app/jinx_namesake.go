package app

import (
	"context"
	"errors"

	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/releaseparse"
)

// titleNamesake reports whether the catalogue holds a show other than show
// whose own title (or original title) is title (resolver/jinx_namesake.go).
func (a *App) titleNamesake(ctx context.Context, show core.RatingKey, title string) (bool, error) {
	if a.catalog == nil {
		return false, errors.New("catalogue unavailable")
	}
	shows, err := a.catalog.SearchAll(ctx, title, []core.MediaType{core.TypeShow}, 20)
	if err != nil {
		return false, err
	}
	for _, t := range shows {
		if t.Key == show {
			continue
		}
		if releaseparse.TitlesMatch(t.Title, title) || (t.OriginalTitle != "" && releaseparse.TitlesMatch(t.OriginalTitle, title)) {
			return true, nil
		}
	}
	return false, nil
}
