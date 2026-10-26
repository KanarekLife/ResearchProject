package cards

import (
	"testing"

	e "mcbench/game/engine"
)

// get_card must survive the ways models actually misspelled names in runs.
func TestSearchToleratesHowModelsTypeNames(t *testing.T) {
	cases := map[string]string{
		"Breakin' &amp; Takin'": "Breakin' & Takin'", // HTML-escaped
		"Breakin & Takin":       "Breakin' & Takin'", // no apostrophes
		"swinging web-kick":     "Swinging Web Kick",
		"Swingin Web Kick":      "Swinging Web Kick", // one typo
		"Horn":                  "Enhanced Ivory Horn",
		"Armored Rhino":         "Armored Rhino Suit",
		"Peter Parker":          "Peter Parker", // an alter-ego side
	}
	for query, want := range cases {
		ds := Search(query)
		if len(ds) != 1 || ds[0].Name != want {
			t.Errorf("Search(%q) = %v, want %s", query, names(ds), want)
		}
	}
}

func TestSearchListsAmbiguousAndRejectsUnknownNames(t *testing.T) {
	if ds := Search("rhino"); len(ds) != 3 {
		t.Errorf("Search(rhino) = %v, want the three villain stages", names(ds))
	}
	if ds := Search(Search("rhino")[1].Code); len(ds) != 1 {
		t.Errorf("a code must pick one stage, got %v", names(ds))
	}
	if ds := Search("Iron Fist"); len(ds) != 0 {
		t.Errorf("Search(Iron Fist) = %v, want nothing", names(ds))
	}
}

func names(ds []*e.CardDef) []string {
	var out []string
	for _, d := range ds {
		out = append(out, d.Name)
	}
	return out
}
