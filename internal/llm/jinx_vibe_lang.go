package llm

import "strings"

// Models answer "Telugu" as often as "te" even when asked for ISO 639-1
// codes; the parser kept only two-letter codes and silently dropped the
// rest (jinx).
var vibeLanguageCodes = map[string]string{
	"english": "en", "hindi": "hi", "telugu": "te", "tamil": "ta", "malayalam": "ml", "kannada": "kn",
	"marathi": "mr", "bengali": "bn", "bangla": "bn", "gujarati": "gu", "punjabi": "pa", "odia": "or", "urdu": "ur",
	"japanese": "ja", "korean": "ko", "chinese": "zh", "mandarin": "zh", "cantonese": "cn", "spanish": "es",
	"french": "fr", "german": "de", "italian": "it", "portuguese": "pt", "russian": "ru", "turkish": "tr",
	"thai": "th", "indonesian": "id", "arabic": "ar", "persian": "fa", "farsi": "fa", "swedish": "sv",
	"danish": "da", "norwegian": "no", "dutch": "nl", "polish": "pl", "hebrew": "he", "greek": "el",
}

// vibeLanguage is l as an ISO 639-1 code, or "".
func vibeLanguage(l string) string {
	l = strings.ToLower(strings.TrimSpace(l))
	if len(l) == 2 {
		return l
	}
	return vibeLanguageCodes[l]
}

// maxVibeLanguages: "South Indian" alone is four (te, ta, ml, kn).
const maxVibeLanguages = 6
