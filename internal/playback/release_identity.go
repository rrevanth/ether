package playback

import (
	"bytes"
	"encoding/json"
	"path"
	"strings"

	"github.com/braedonsaunders/ether/internal/core"
	"github.com/braedonsaunders/ether/internal/releaseparse"
)

const debridPayloadPrefix = "ETHER-DEBRID/1\n"

// payloadMatchesTitle verifies that a cached or resumed debrid payload's
// chosen file path matches the requested episode.
func payloadMatchesTitle(nzb []byte, title *core.Title) bool {
	if title == nil || title.Type != core.TypeEpisode || !bytes.HasPrefix(nzb, []byte(debridPayloadPrefix)) {
		return true
	}
	var p struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(nzb[len(debridPayloadPrefix):], &p); err != nil || p.Path == "" {
		return true
	}
	season, episode := title.Season, title.Episode
	if title.Key.Kind() == core.KindEpisode {
		if season == 0 {
			season = title.Key.Season()
		}
		if episode == 0 {
			episode = title.Key.Episode()
		}
	}
	if episode <= 0 {
		return true
	}
	normPath := strings.ReplaceAll(p.Path, "\\", "/")
	baseInfo := releaseparse.Parse(path.Base(normPath))
	if season == 0 {
		// jinx: a special's file was picked by debrid ChooseFile: by the
		// add-on's own file pointer, whose numbering is the add-on's (IMDb's
		// "S00E14" for TMDB's S00E05), or by an explicit S00Exx. Neither its
		// number nor a folder's seasonless range ("0001-1071") can judge it.
		return true
	}
	fullInfo := releaseparse.Parse(strings.ReplaceAll(normPath, "/", " "))
	if len(baseInfo.Seasons) == 0 && len(fullInfo.Seasons) > 0 {
		baseInfo.Seasons = fullInfo.Seasons
	}
	if len(baseInfo.Episodes) > 0 {
		return episodeIdentity(baseInfo, season, episode, title.AbsoluteEpisode)
	}
	if len(fullInfo.Episodes) > 0 {
		return episodeIdentity(fullInfo, season, episode, title.AbsoluteEpisode)
	}
	return true
}

// episodeIdentity is the resolver's episode match (resolver.go
// episodeMatchesRequest): the season's episode, or the absolute number a
// long anime names ("S07E12 (0207)", "S01E1121") (jinx).
func episodeIdentity(info core.ReleaseInfo, season, episode, absolute int) bool {
	if releaseparse.EpisodeMatches(info, season, episode) || releaseparse.EpisodeMatchesAbsolute(info, absolute) {
		return true
	}
	return absolute > 0 && len(info.Seasons) == 1 && len(info.Episodes) == 1 && !info.FullSeason &&
		info.Episodes[0] == absolute && (info.Seasons[0] == 1 || absolute >= 100)
}

// releaseMatchesTitle applies the same episode identity constraints to fresh,
// warmed, remembered and failover choices. Unknown filename metadata is not
// proof of identity; it only permits the resolver/file selector to continue.
func releaseMatchesTitle(rel core.Release, title *core.Title) bool {
	if title == nil {
		return true
	}
	expected := title.IDs
	if title.Type == core.TypeEpisode {
		expected = title.ShowIDs
		if expected.TMDB == 0 && title.Key.Kind() == core.KindEpisode {
			expected.TMDB = title.Key.TMDB()
		}
	}
	idMatch, idConflict := releaseparse.CompareTitleIDs(rel.TitleIDs, expected)
	if idConflict {
		return false
	}
	if title.Type != core.TypeEpisode {
		return true
	}
	season, episode := title.Season, title.Episode
	if title.Key.Kind() == core.KindEpisode {
		if season == 0 {
			season = title.Key.Season()
		}
		if episode == 0 {
			episode = title.Key.Episode()
		}
	}
	info := rel.Info
	if len(info.Episodes) == 0 && !info.FullSeason {
		info = releaseparse.Parse(rel.Name)
	}
	if episode > 0 && (len(info.Episodes) > 0 || info.FullSeason) {
		if info.FullSeason {
			if len(info.Seasons) > 0 {
				found := false
				for _, n := range info.Seasons {
					found = found || n == season
				}
				if !found {
					return false
				}
			}
		} else if !episodeIdentity(info, season, episode, title.AbsoluteEpisode) {
			return false
		}
	}
	if rel.Target != nil && episode > 0 &&
		(rel.Target.Episode != episode || rel.Target.Season != season) {
		return false
	}
	year := releaseparse.EpisodeSeriesYear(rel.Name, rel.Info.Year)
	if !releaseparse.SeriesYearMatches(year, title.ShowYear) ||
		(year == 0 && !idMatch && !(season == 0 && episode > 0) && !releaseparse.EpisodeTitleMatches(rel.Name, title.Title)) { // jinx: specials' titles are judged by the resolver (jinx_specials.go)
		return false
	}
	// A torrent descriptor embeds the expected identity for file selection.
	// Older cached descriptors must be resolved again to acquire that target.
	return rel.Protocol != core.ProtocolTorrent || title.ShowYear <= 0 ||
		(rel.Target != nil && rel.Target.ShowYear == title.ShowYear && rel.Target.EpisodeTitle == title.Title)
}
