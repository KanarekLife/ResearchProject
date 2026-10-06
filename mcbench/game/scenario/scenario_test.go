package scenario_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mcbench/game/scenario"
)

// workRoot copies the real data into a temp dir and writes one scenario file,
// so a test can break one thing at a time.
func workRoot(t *testing.T, scenarioYAML string) (path, root string) {
	t.Helper()
	root = t.TempDir()
	if err := os.CopyFS(root, os.DirFS("../../data")); err != nil {
		t.Fatal(err)
	}
	path = filepath.Join(root, "bad.yaml")
	if err := os.WriteFile(path, []byte(scenarioYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	return path, root
}

func TestLoadRejectsBrokenScenarios(t *testing.T) {
	const ok = "deck: spider-man-justice\nvillain: rhino\nencounter_sets: [standard]\nseeds: [1]\n"
	for name, tc := range map[string]struct{ yaml, want string }{
		"no seeds":      {"deck: spider-man-justice\nvillain: rhino\n", "no seeds"},
		"missing deck":  {strings.Replace(ok, "spider-man-justice", "no-such-deck", 1), "no-such-deck"},
		"missing set":   {strings.Replace(ok, "[standard]", "[no-such-set]", 1), "no-such-set"},
		"unknown field": {ok + "bogus: 1\n", "bogus"},
	} {
		path, root := workRoot(t, tc.yaml)
		_, err := scenario.Load(path, root)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %v, want one mentioning %q", name, err, tc.want)
		}
	}
}

func TestLoadUnknownCard(t *testing.T) {
	path, root := workRoot(t, "deck: spider-man-justice\nvillain: rhino\nseeds: [1]\n")
	deck := filepath.Join(root, "decks", "spider-man-justice.yaml")
	b, err := os.ReadFile(deck)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(b), "obligation:", "  - No Such Card\nobligation:", 1)
	if err := os.WriteFile(deck, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := scenario.Load(path, root); err == nil || !strings.Contains(err.Error(), "No Such Card") {
		t.Errorf("error %v, want one naming the unknown card", err)
	}
}

func TestLoadDefaultsAndID(t *testing.T) {
	path, root := workRoot(t, "deck: spider-man-justice\nvillain: rhino\nseeds: [7]\n")
	sc, err := scenario.Load(path, root)
	if err != nil {
		t.Fatal(err)
	}
	if sc.ID != "bad" || sc.MaxRounds != 20 || sc.MaxDecisions != 1000 {
		t.Errorf("id=%q rounds=%d decisions=%d, want the file name and defaults 20/1000", sc.ID, sc.MaxRounds, sc.MaxDecisions)
	}
}
