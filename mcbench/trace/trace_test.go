package trace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeTrace writes lines to a trace file, newline-terminated as session.Trace
// writes them, and returns its path.
func writeTrace(t *testing.T, lines ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "game.jsonl")
	body := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

const (
	startLine  = `{"type":"start","time":"t0","scenario":"s","seed":1,"sample":0,"player":"model:x","instructions":{"names":["rules"],"sha256":"abc"}}`
	choiceLine = `{"type":"choice","time":"t1","decision":1,"choice":{"round":1,"kind":"mulligan","key":"mulligan:keep","reasoning":"good hand"},"text":"Keep this hand","events":["Setup"],"status":"awaiting_decision"}`
	replyLine  = `{"type":"message","time":"t2","decision":1,"message":{"role":"assistant","reasoning_content":"thinking hard","tool_calls":[{"function":{"name":"choose_option"}}]},"prompt_tokens":100,"completion_tokens":20}`
	endLine    = `{"type":"end","time":"t3","status":"lost","rounds":2,"score":0.1,"usage":{"requests":3,"input_tokens":500,"output_tokens":60}}`
)

func TestDecodeSkipsPartialTrailingLine(t *testing.T) {
	path := writeTrace(t, startLine, choiceLine, replyLine, endLine)
	appendFile(t, path, `{"type":"choi`) // a line still being written
	entries, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 {
		t.Fatalf("got %d entries, want 4 (the partial line dropped)", len(entries))
	}
	if !IsChoice(entries[1]) {
		t.Errorf("entry 1 is %q, want a choice", entries[1].Type)
	}
	if !IsEnd(entries[3]) {
		t.Errorf("entry 3 is %q, want the end", entries[3].Type)
	}
}

func TestDecodeFindsModelCallsAndReasoning(t *testing.T) {
	entries, err := ReadFile(writeTrace(t, replyLine))
	if err != nil {
		t.Fatal(err)
	}
	e := entries[0]
	if got := len(e.Message.ToolCalls); got != 1 {
		t.Fatalf("got %d tool calls, want 1", got)
	}
	if name := e.Message.ToolCalls[0].Function.Name; name != "choose_option" {
		t.Errorf("got tool %q, want choose_option", name)
	}
	if e.Message.Reasoning != "thinking hard" {
		t.Errorf("got reasoning %q", e.Message.Reasoning)
	}
	if e.PromptTokens != 100 || e.CompletionTokens != 20 {
		t.Errorf("got %d/%d tokens, want 100/20", e.PromptTokens, e.CompletionTokens)
	}
}

// TestFollowerReadsAppends covers the live case: entries appear over time, and a
// line may be half-written when the follower looks.
func TestFollowerReadsAppends(t *testing.T) {
	path := writeTrace(t, startLine)
	f, err := Follow(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	got := func() []Entry {
		t.Helper()
		es, err := f.Next(time.Second)
		if err != nil {
			t.Fatal(err)
		}
		return es
	}
	if es := got(); len(es) != 1 || es[0].Type != "start" {
		t.Fatalf("first read got %d entries, want the start line", len(es))
	}

	appendFile(t, path, choiceLine+"\n")
	if es := got(); len(es) != 1 || !IsChoice(es[0]) {
		t.Fatalf("second read got %d entries, want the choice", len(es))
	}

	// A line that is not newline-terminated yet must not be decoded or lost.
	appendFile(t, path, `{"type":"mes`)
	if es := got(); len(es) != 0 {
		t.Fatalf("read %d entries from a partial line, want 0", len(es))
	}
	appendFile(t, path, `sage","decision":1,"prompt_tokens":10,"completion_tokens":2}`+"\n")
	es := got()
	if len(es) != 1 || es[0].Type != "message" {
		t.Fatalf("got %d entries, want the completed message", len(es))
	}
	if es[0].PromptTokens != 10 {
		t.Errorf("got prompt_tokens %d, want 10", es[0].PromptTokens)
	}

	appendFile(t, path, endLine+"\n")
	if es := got(); len(es) != 1 || !IsEnd(es[0]) {
		t.Fatalf("got %d entries, want the end", len(es))
	}
}

func TestFollowerReturnsNilWhenFileIdle(t *testing.T) {
	f, err := Follow(writeTrace(t, startLine))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.Next(10 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	es, err := f.Next(10 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if es != nil {
		t.Errorf("got %d entries from an idle file, want none", len(es))
	}
}

func TestPickRunAndGame(t *testing.T) {
	dir := t.TempDir()
	for _, run := range []string{"20260101T000000Z-model_a", "20260102T000000Z-model_b"} {
		live := filepath.Join(dir, run, "live")
		if err := os.MkdirAll(live, 0o755); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"sc_seed1_s0", "sc_seed2_s0"} {
			if err := os.WriteFile(filepath.Join(live, name+".jsonl"), []byte(startLine+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	runs, err := Discover(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 2 {
		t.Fatalf("got %d runs, want 2", len(runs))
	}
	if runs[0].ID != "20260102T000000Z-model_b" {
		t.Errorf("runs are not newest first: got %q first", runs[0].ID)
	}
	// An empty selector picks the newest run.
	newest, err := PickRun(runs, "")
	if err != nil || newest.ID != runs[0].ID {
		t.Errorf("PickRun(\"\") = %q, %v; want the newest run", newest.ID, err)
	}
	// A prefix resolves when it is unique, and is an error when it is not.
	if _, err := PickRun(runs, "20260101"); err != nil {
		t.Errorf("unique prefix failed: %v", err)
	}
	if _, err := PickRun(runs, "2026"); err == nil {
		t.Error("an ambiguous prefix should be an error")
	}
	// An empty game selector picks the last game.
	last, err := PickGame(newest, "")
	if err != nil || last.Name != "sc_seed2_s0" {
		t.Errorf("PickGame(\"\") = %q, %v; want the last game", last.Name, err)
	}
	if _, err := PickGame(newest, "seed1"); err != nil {
		t.Errorf("unique game substring failed: %v", err)
	}
	if _, err := PickGame(newest, "nope"); err == nil {
		t.Error("an unknown game should be an error")
	}
}

func appendFile(t *testing.T, path, s string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(s); err != nil {
		t.Fatal(err)
	}
}
