package llm

import (
	"slices"
	"testing"
)

func TestVibeLanguagesAcceptNamesAndSouthIndiaFits(t *testing.T) {
	v, ok := parseVibeJSON(`{"vibe":true,"title":"South Indian Thrillers","type":"movie","genres":["Thriller"],"yearFrom":0,"yearTo":0,"minRating":7,"maxRuntimeMin":0,"minRuntimeMin":0,"languages":["Telugu","Tamil","ml","Kannada","Klingon"],"countries":["India"],"personName":"","studio":"","similarTo":"","sort":"rating"}`)
	if !ok {
		t.Fatal("not parsed")
	}
	if !slices.Equal(v.Filter.Languages, []string{"te", "ta", "ml", "kn"}) {
		t.Fatalf("languages %v", v.Filter.Languages)
	}
}
