package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Instructions are the documents given to a model agent. The instruction set
// is the main independent variable of the benchmark, so its content hash is
// recorded with every run.
type Instructions struct {
	Names  []string `json:"names"`
	SHA256 string   `json:"sha256"`
	Text   string   `json:"-"`
}

// LoadInstructions concatenates dir/<name>.md for each name, in order.
func LoadInstructions(dir string, names []string) (Instructions, error) {
	var parts []string
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, n+".md"))
		if err != nil {
			return Instructions{}, fmt.Errorf("instructions %q: %w", n, err)
		}
		parts = append(parts, strings.TrimSpace(string(data)))
	}
	text := strings.Join(parts, "\n\n---\n\n")
	sum := sha256.Sum256([]byte(text))
	return Instructions{Names: names, SHA256: hex.EncodeToString(sum[:])[:16], Text: text}, nil
}
