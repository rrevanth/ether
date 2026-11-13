package llm

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/braedonsaunders/ether/internal/core"
)

// Natural ("AI") search asked the model to "suggest at most N" and the model
// stopped at about eight, some of which the catalogue then could not match
// or which broke the request's language ("high rated South Indian thrillers"
// gave Kahaani, a Hindi film, and Drishyam's Hindi remake). The model is now
// asked for a full list, over-asked so catalogue misses do not shorten it,
// held to the request's languages and regions, and given time to answer
// (jinx).

const (
	naturalSearchWait    = 10 * time.Second // first run only; repeats come from the cache
	naturalSearchMinAsk  = 25
	naturalSearchMaxAsk  = 50
	naturalSearchOverAsk = 10
)

// naturalSearchAsk is how many titles to ask for when the caller keeps limit.
func naturalSearchAsk(limit int) int {
	return min(max(limit+naturalSearchOverAsk, naturalSearchMinAsk), naturalSearchMaxAsk)
}

func naturalSearchPrompt(query string, limit int) string {
	return fmt.Sprintf(`Find real titles matching this request: %q.
Suggest %d titles, best matches first; return fewer only if that many real titles matching every part of the request do not exist.
Every title must satisfy all of the request's constraints: genre, era, rating, mood, and especially language and region. A language or region means films made in it (original language), never remakes or dubs from elsewhere; "South Indian" means Telugu, Tamil, Malayalam or Kannada films, "Bollywood" means Hindi films.
Give each title's original release year exactly. Reply JSON {"titles":[{"title":"...","year":2000,"type":"movie"}]}.`, query, naturalSearchAsk(limit))
}

var reWord = regexp.MustCompile(`[a-z]+`)

// requestLanguages are the original languages a natural search names
// ("telugu", "malayalam", "south indian", "korean"), as ISO 639-1 codes.
// The model is told to respect them but does not always: the catalogue's
// language decides.
func requestLanguages(query string) []string {
	q := strings.ToLower(query)
	var out []string
	add := func(codes ...string) {
		for _, c := range codes {
			if !slices.Contains(out, c) {
				out = append(out, c)
			}
		}
	}
	if strings.Contains(q, "south indian") || strings.Contains(q, "south-indian") || strings.Contains(q, "dravidian") {
		add("te", "ta", "ml", "kn")
	}
	if strings.Contains(q, "bollywood") {
		add("hi")
	}
	if strings.Contains(q, "tollywood") {
		add("te")
	}
	if strings.Contains(q, "kollywood") {
		add("ta")
	}
	if strings.Contains(q, "mollywood") {
		add("ml")
	}
	if strings.Contains(q, "sandalwood") {
		add("kn")
	}
	if strings.Contains(q, "k-drama") || strings.Contains(q, "kdrama") {
		add("ko")
	}
	for _, w := range reWord.FindAllString(q, -1) {
		if w == "english" || w == "indian" {
			continue // "English subtitles", "Indian" spans every language
		}
		if code, ok := vibeLanguageCodes[w]; ok {
			add(code)
		}
	}
	return out
}

// inRequestLanguages keeps the titles whose original language the request
// names; titles of unknown language stay. Without named languages, all stay.
func inRequestLanguages(query string, titles []core.Title) []core.Title {
	langs := requestLanguages(query)
	if len(langs) == 0 {
		return titles
	}
	out := titles[:0:0]
	for _, t := range titles {
		if t.Language == "" || slices.Contains(langs, strings.ToLower(t.Language)) {
			out = append(out, t)
		}
	}
	return out
}
