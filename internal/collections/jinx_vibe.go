package collections

import (
	"context"
	"time"

	"github.com/braedonsaunders/ether/internal/core"
)

// Vibe searches aim for a full row (jinx). The LLM refiner gets time to
// answer (a hosted model takes 1-2 s; at 1.2 s the search fell back to plain
// title search), ratings only count with enough votes, and a filter that
// finds fewer than vibeMinResults relaxes its rating floor, then its vote
// floor, before giving up ("high rated South Indian thrillers" found 8 titles
// at rating 8 and 44 at 7).

const (
	vibeRefineTimeout = 4 * time.Second
	vibeMinResults    = 20
	vibeVoteFloor     = 20
	vibeLowVoteFloor  = 10
	vibeRatingFloor   = 6.5
)

// withVoteFloor makes a rating filter or rating sort count only titles
// enough people rated (a 9.0 from three votes is no rating).
func withVoteFloor(v *core.VibeQuery) {
	if (v.Filter.MinRating > 0 || v.Sort == core.SortRating) && v.Filter.MinVotes == 0 && v.SimilarKey == 0 {
		v.Filter.MinVotes = vibeVoteFloor
	}
}

// relaxedVibePage lowers e's rating floor in half steps (not below
// vibeRatingFloor), then its vote floor, until the page holds
// vibeMinResults titles; e keeps the filter that produced the page.
func (s *Service) relaxedVibePage(ctx context.Context, user *core.User, e *ephemeralEntry, page core.Page) (core.Page, error) {
	if e.vibe.SimilarKey != 0 {
		return page, nil
	}
	try := func(f core.Filter) (bool, error) {
		saved := e.vibe.Filter
		e.vibe.Filter = f
		p, err := s.ephemeralPage(ctx, user, e, 0, vibeTopItems)
		if err != nil {
			e.vibe.Filter = saved
			return false, err
		}
		if p.Total <= page.Total {
			e.vibe.Filter = saved
			return false, nil
		}
		page = p
		return p.Total >= vibeMinResults, nil
	}
	for page.Total < vibeMinResults && e.vibe.Filter.MinRating > vibeRatingFloor {
		f := e.vibe.Filter
		f.MinRating = max(f.MinRating-0.5, vibeRatingFloor)
		if done, err := try(f); err != nil || done {
			return page, err
		}
		e.vibe.Filter.MinRating = f.MinRating // keep stepping even when a step adds nothing
	}
	for _, votes := range []int{vibeLowVoteFloor, 0} {
		if page.Total >= vibeMinResults || e.vibe.Filter.MinVotes <= votes {
			continue
		}
		f := e.vibe.Filter
		f.MinVotes = votes // last: no floor, never fewer titles than before the floor existed
		if _, err := try(f); err != nil {
			return page, err
		}
	}
	return page, nil
}
