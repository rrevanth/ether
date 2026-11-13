package llm

import (
	"slices"
	"strings"
	"testing"

	"github.com/braedonsaunders/ether/internal/core"
)

func TestNaturalSearchAsksForAFullListHeldToTheRequest(t *testing.T) {
	for limit, want := range map[int]int{5: 25, 20: 30, 30: 40, 50: 50} {
		if got := naturalSearchAsk(limit); got != want {
			t.Errorf("naturalSearchAsk(%d) = %d, want %d", limit, got, want)
		}
	}
	p := naturalSearchPrompt("high rated south indian thriller movies", 20)
	for _, must := range []string{"Suggest 30 titles", "original language", "Telugu, Tamil, Malayalam or Kannada"} {
		if !strings.Contains(p, must) {
			t.Errorf("prompt lacks %q:\n%s", must, p)
		}
	}
	if strings.Contains(p, "at most") {
		t.Error("prompt still invites a short list")
	}
}

func TestNaturalSearchKeepsTheRequestsLanguages(t *testing.T) {
	cases := map[string][]string{
		"high rated south indian thriller movies": {"te", "ta", "ml", "kn"},
		"feel good malayalam movies":              {"ml"},
		"korean crime thrillers":                  {"ko"},
		"indian movies with english subtitles":    nil,
		"90s action movies":                       nil,
	}
	for q, want := range cases {
		if got := requestLanguages(q); !slices.Equal(got, want) {
			t.Errorf("requestLanguages(%q) = %v, want %v", q, got, want)
		}
	}
	titles := []core.Title{{Title: "Drishyam", Language: "ml"}, {Title: "Andhadhun", Language: "hi"}, {Title: "Ratsasan", Language: "ta"}, {Title: "Unknown"}}
	got := inRequestLanguages("high rated south indian thriller movies", titles)
	var names []string
	for _, t := range got {
		names = append(names, t.Title)
	}
	if !slices.Equal(names, []string{"Drishyam", "Ratsasan", "Unknown"}) {
		t.Fatalf("kept %v", names)
	}
}
