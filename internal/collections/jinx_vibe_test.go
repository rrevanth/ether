package collections_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/braedonsaunders/ether/internal/catalog"
	"github.com/braedonsaunders/ether/internal/collections"
	"github.com/braedonsaunders/ether/internal/config"
	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/store"
)

type fixedRefiner struct{ v core.VibeQuery }

func (r fixedRefiner) RefineVibe(context.Context, *core.User, string) (core.VibeQuery, bool, error) {
	return r.v, true, nil
}

// "High rated South Indian thrillers": the model asks for rating 8, which
// three titles reach; the row relaxes to 20+ without a two-vote 9.5 on top.
func TestVibeRelaxesToAFullRowAndIgnoresRatingsFewVotesGave(t *testing.T) {
	ctx := context.Background()
	db, err := store.OpenMemory(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	cat, err := catalog.New(db, nil, config.Catalog{Language: "en-US", Region: "US"}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	add := func(id int64, rating float64, votes int) {
		title := core.Title{Key: core.MovieKey(id), Type: core.TypeMovie, IDs: core.IDs{TMDB: id}, Title: fmt.Sprintf("Thriller %d", id),
			Genres: []string{"Thriller"}, Language: "te", Year: 2018, Released: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			Rating: rating, VoteCount: votes, Popularity: 10}
		if err := cat.UpsertTitle(ctx, title, true); err != nil {
			t.Fatal(err)
		}
	}
	for i := int64(1); i <= 3; i++ {
		add(i, 8.2, 150)
	}
	for i := int64(10); i < 35; i++ {
		add(i, 7.2, 100)
	}
	add(99, 9.5, 2) // two votes
	ref := fixedRefiner{core.VibeQuery{Title: "High Rated South Indian Thrillers", Sort: core.SortRating, Desc: true,
		Filter: core.Filter{Type: core.TypeMovie, Genres: []string{"Thriller"}, Languages: []string{"te", "ta", "ml", "kn"}, MinRating: 8}}}
	svc, err := collections.New(db, cat, slog.Default(), collections.WithIndexTTL(0), collections.WithLiveTTL(0), collections.WithVibeRefiner(ref))
	if err != nil {
		t.Fatal(err)
	}
	c, items, ok, err := svc.VibeCollection(ctx, &core.User{ID: 7}, "high rated south indian thriller movies")
	if err != nil || !ok {
		t.Fatalf("VibeCollection: ok=%v err=%v", ok, err)
	}
	if c.ChildCount < 20 || len(items) < 20 {
		t.Fatalf("row of %d (%d items), want 20+", c.ChildCount, len(items))
	}
	for _, it := range items {
		if it.Key == core.MovieKey(99) {
			t.Fatal("a two-vote 9.5 made the row")
		}
	}
	if items[0].Rating < 8 {
		t.Fatalf("best-rated first: got %v", items[0].Rating)
	}
}
