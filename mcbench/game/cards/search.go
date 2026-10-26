package cards

import (
	"html"
	"strings"
	"unicode"

	e "mcbench/game/engine"
)

// Search resolves a card name as a model or a person might type it. It tries,
// in order: a code; the name ignoring case, spacing, punctuation and HTML
// escapes ("Breakin' &amp; Takin'", "breakin takin"); a part of a name
// ("Horn"); a name within a few typos. It returns every card side
// that matches at the first level that matches anything, sorted by code, so
// several results mean the name is ambiguous. Alter-ego sides are included.
func Search(query string) []*e.CardDef {
	if d, ok := byCode[strings.TrimSpace(query)]; ok {
		return []*e.CardDef{d}
	}
	key := normalize(query)
	if key == "" {
		return nil
	}
	levels := []func(name string) bool{
		func(name string) bool { return name == key },
		func(name string) bool { return len(key) >= minPartial && strings.Contains(name, key) },
		func(name string) bool { return editDistance(name, key) <= len(key)/charsPerTypo },
	}
	all := All()
	for _, match := range levels {
		var out []*e.CardDef
		for _, d := range all {
			if match(normalize(d.Name)) {
				out = append(out, d)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return nil
}

const (
	// minPartial is the shortest query matched as part of a name, so that
	// "a" does not match every card.
	minPartial = 4
	// charsPerTypo allows one typo per this many characters of the query:
	// "swingin web kick" finds Swinging Web Kick, "Iron Fist" does not
	// find Iron Man.
	charsPerTypo = 8
)

// normalize keeps only the lower-cased letters and digits of an unescaped
// name.
func normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(html.UnescapeString(s)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// editDistance is the Levenshtein distance between a and b.
func editDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur := make([]int, len(rb)+1)
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(rb)]
}
