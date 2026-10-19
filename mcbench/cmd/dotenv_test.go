package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(`
# comment
PLAIN=value
export QUOTED="quoted value"
SINGLE='single value'
`), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PLAIN", "existing")
	if err := loadDotEnv(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("PLAIN"); got != "existing" {
		t.Errorf("PLAIN = %q, want existing", got)
	}
	if got := os.Getenv("QUOTED"); got != "quoted value" {
		t.Errorf("QUOTED = %q, want quoted value", got)
	}
	if got := os.Getenv("SINGLE"); got != "single value" {
		t.Errorf("SINGLE = %q, want single value", got)
	}
}

func TestLoadDotEnvMissingIsFine(t *testing.T) {
	if err := loadDotEnv(filepath.Join(t.TempDir(), ".env")); err != nil {
		t.Fatal(err)
	}
}

func TestLoadDotEnvRejectsMalformedEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("not an assignment\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := loadDotEnv(path); err == nil {
		t.Fatal("malformed entry accepted")
	}
}
