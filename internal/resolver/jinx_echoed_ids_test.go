package resolver

import (
	"strings"
	"testing"

	"github.com/braedonsaunders/ether/internal/core"
)

// Asked by Money Heist's IMDb id, an add-on answered with "La casa di Davide"
// (House of David). The id it carries is only the query's: the title check
// still applies. An id the source itself reported still vouches.
func TestEchoedIDsDoNotVouchForAnotherTitle(t *testing.T) {
	ids := core.IDs{IMDB: "tt6468322", TMDB: 71446}
	title := &core.Title{Type: core.TypeEpisode, Title: "Episode 1", ShowTitle: "Money Heist", ShowOriginalTitle: "La casa de papel", ShowYear: 2017, Season: 1, Episode: 1}
	sreq := core.SearchRequest{Type: core.TypeEpisode, Title: "Money Heist", Year: 2017, Season: 1, Episode: 1, IDs: ids}
	echoed := withParsed(rel("a", "La.casa.di.Davide.S01E01.Un.pastore.e.un.re..2160p.AMZN.WEB-DL.DDP5.1.DV.HDR.H265-G66", 6000))
	echoed.TitleIDs, echoed.TitleIDsEchoed = ids, true
	if got := rejectIdentityYears(echoed, title, sreq, "en", nil); !strings.HasPrefix(got, "title mismatch") {
		t.Fatalf("echoed id let another show through: %q", got)
	}
	reported := echoed
	reported.TitleIDsEchoed = false
	if got := rejectIdentityYears(reported, title, sreq, "en", nil); got != "" {
		t.Fatalf("a reported id no longer vouches: %q", got)
	}
	own := withParsed(rel("b", "La.Casa.De.Papel.S01E01.1080p.NF.WEB-DL", 1500))
	own.TitleIDs, own.TitleIDsEchoed = ids, true
	if got := rejectIdentityYears(own, title, sreq, "en", nil); got != "" {
		t.Fatalf("the show's own release rejected: %q", got)
	}
}
