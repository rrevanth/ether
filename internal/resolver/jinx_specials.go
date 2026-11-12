package resolver

import (
	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/releaseparse"
)

// rejectSpecial: release groups number specials their own way (One Piece:
// A&C's "S00E02 - Movie 01", CR's "S00E01 One Piece Heroines"), so a release
// for a season-0 episode must say it is a special (season 0 or a
// Specials/OVA/OAD word; a regular "Ep.001-165" batch is not) and a title
// after its S00Exx must agree with TMDB's (releaseparse.SpecialTitleAgrees).
// This replaces upstream's episode-title rule for season 0, which rejected
// partial titles ("Episode of Luffy") and let multi-language releases through
// unchecked (jinx).
func rejectSpecial(rel core.Release, sreq core.SearchRequest) string {
	if sreq.Type != core.TypeEpisode || sreq.Season != 0 || sreq.Episode <= 0 {
		return ""
	}
	marked := releaseparse.NamesSpecial(rel.Name)
	for _, s := range rel.Info.Seasons {
		marked = marked || s == 0
	}
	if !marked {
		return "names no special"
	}
	shows := append([]string{sreq.Title, sreq.OriginalTitle}, sreq.AlternativeTitles...)
	if !releaseparse.SpecialTitleAgrees(rel.Name, sreq.EpisodeTitle, shows...) {
		return "special's title differs"
	}
	return ""
}
