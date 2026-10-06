package boardgui

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"mcbench/benchmark"
	"mcbench/constants"
	"mcbench/game/scenario"
	"mcbench/player"
	"mcbench/player/heuristic"
	"mcbench/player/instruction"
	"mcbench/trace"
)

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	os.Exit(m.Run())
}

// playTrace plays one scripted game into a trace file and returns its path and
// record.
func playTrace(t *testing.T, scs []*scenario.Scenario, p player.Player, seed uint64, sample int) (string, benchmark.Record) {
	t.Helper()
	dir := t.TempDir()
	rec := benchmark.Play(context.Background(), scs[0], p, instruction.Instructions{}, seed, sample, dir, benchmark.DefaultWeights)
	paths, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if len(paths) != 1 {
		t.Fatalf("want one trace, got %v", paths)
	}
	return paths[0], rec
}

func loadScenarios(t *testing.T) []*scenario.Scenario {
	t.Helper()
	scs, err := scenario.LoadDir("../data/scenarios", "../data")
	if err != nil {
		t.Fatal(err)
	}
	return scs
}

// oneTrace is a short scripted game for tests that need any trace.
func oneTrace(t *testing.T) ([]*scenario.Scenario, string, benchmark.Record) {
	scs := loadScenarios(t)
	path, rec := playTrace(t, scs, heuristic.Random{Seed: 7}, scs[0].Seeds[0], 1)
	return scs, path, rec
}

// Every seed replays to one step per decision plus the final board, ending
// where the recorded game ended. Several samples make some games discard one
// of two copies of a card, whose choices share a key.
func TestReplayMatchesTrace(t *testing.T) {
	scs := loadScenarios(t)
	for _, seed := range scs[0].Seeds {
		for _, p := range []player.Player{heuristic.First{}, heuristic.Random{Seed: seed}} {
			for sample := range 3 {
				path, rec := playTrace(t, scs, p, seed, sample)
				entries, err := trace.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				r := newReplay(scs)
				r.apply(entries)
				checkReplay(t, r.snapshot(0), rec)
			}
		}
	}
}

func checkReplay(t *testing.T, snap Snapshot, rec benchmark.Record) {
	t.Helper()
	if snap.Error != "" {
		t.Fatalf("seed %d %s sample %d: %s", rec.Seed, rec.Player, rec.Sample, snap.Error)
	}
	if snap.Total != rec.Decisions+1 {
		t.Fatalf("got %d steps for %d decisions", snap.Total, rec.Decisions)
	}
	for i, s := range snap.Steps[:len(snap.Steps)-1] {
		if s.Decision == nil || s.Choice == nil {
			t.Fatalf("step %d lacks its decision or choice", i)
		}
	}
	last := snap.Steps[len(snap.Steps)-1]
	if last.Decision != nil || last.Status != rec.Status || snap.End == nil || snap.End.Status != rec.Status {
		t.Fatalf("final step %q, end %+v; recorded %q", last.Status, snap.End, rec.Status)
	}
}

// A replay that no longer matches the trace stops with an error and keeps the
// boards it had.
func TestReplayDetectsDivergence(t *testing.T) {
	scs, path, _ := oneTrace(t)
	entries, err := trace.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	bad := 0
	for i, e := range entries {
		if e.Type == constants.TraceChoice {
			if bad++; bad == 3 {
				entries[i].Choice.Key = "no-such-option"
				break
			}
		}
	}
	r := newReplay(scs)
	r.apply(entries)
	snap := r.snapshot(0)
	if snap.Error == "" || snap.Total != 3 {
		t.Fatalf("want an error after 3 steps, got %d steps, error %q", snap.Total, snap.Error)
	}
}

// Following a trace that is still being written picks up new lines, including
// one that was half written on the first read.
func TestReplayFollowsGrowingTrace(t *testing.T) {
	scs, path, rec := oneTrace(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	live := filepath.Join(t.TempDir(), "live.jsonl")
	cut := bytes.Index(data, []byte(`"type":"choice"`)) + 5 // inside the first choice line
	if err := os.WriteFile(live, data[:cut], 0o644); err != nil {
		t.Fatal(err)
	}
	r := newReplay(scs)
	go r.follow(live, 10*time.Millisecond)
	waitFor(t, func() bool { return r.snapshot(0).Total == 1 })

	f, err := os.OpenFile(live, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	f.Write(data[cut:])
	f.Close()
	waitFor(t, r.done)
	if snap := r.snapshot(0); snap.Error != "" || snap.Total != rec.Decisions+1 {
		t.Fatalf("got %d steps, error %q; want %d", snap.Total, snap.Error, rec.Decisions+1)
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(5 * time.Millisecond) {
		if cond() {
			return
		}
	}
	t.Fatal("timed out")
}

// runDir puts traces into a results directory as one run's games and returns
// the results directory.
func runDir(t *testing.T, traces map[string][]byte) string {
	t.Helper()
	results := t.TempDir()
	live := filepath.Join(results, testRun, "live")
	if err := os.MkdirAll(live, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, data := range traces {
		if err := os.WriteFile(filepath.Join(live, name+".jsonl"), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return results
}

const testRun = "20260101T000000Z-test"

// finished fetches the steps of a finished game once its replay has read the
// whole trace.
func finished(t *testing.T, url string) Snapshot {
	t.Helper()
	var snap Snapshot
	waitFor(t, func() bool {
		snap = Snapshot{}
		return getJSON(t, url, &snap) == http.StatusOK && snap.End != nil
	})
	return snap
}

func getJSON(t *testing.T, url string, v any) int {
	t.Helper()
	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		if err := json.NewDecoder(res.Body).Decode(v); err != nil {
			t.Fatal(err)
		}
	}
	return res.StatusCode
}

// The server only reads: it serves the page, the runs and the steps, and
// refuses any other method.
func TestServerIsReadOnly(t *testing.T) {
	scs, path, _ := oneTrace(t)
	data, _ := os.ReadFile(path)
	results := runDir(t, map[string][]byte{"a": data})
	srv := httptest.NewServer(newServer(newViewer(results, testRun, "a", scs)))
	defer srv.Close()

	snap := finished(t, srv.URL+"/api/steps?run="+testRun+"&game=a&from=1")
	if snap.From != 1 || len(snap.Steps) != snap.Total-1 {
		t.Fatalf("steps: from %d, %d of %d", snap.From, len(snap.Steps), snap.Total)
	}

	res, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET /: %s", res.Status)
	}

	for _, route := range []string{"/", "/api/steps", "/api/cards", "/api/runs"} {
		res, err := http.Post(srv.URL+route, "application/json", strings.NewReader("{}"))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != http.StatusMethodNotAllowed {
			t.Fatalf("POST %s: %s", route, res.Status)
		}
	}
}

// The runs list every game with its state, a game is chosen by run and name,
// and nothing outside the listing is served.
func TestServerSelectsListedGames(t *testing.T) {
	scs, path, rec := oneTrace(t)
	data, _ := os.ReadFile(path)
	results := runDir(t, map[string][]byte{"a": data, "b": data[:bytes.IndexByte(data, '\n')+1]})
	os.WriteFile(filepath.Join(filepath.Dir(results), "outside.jsonl"), data, 0o644)
	srv := httptest.NewServer(newServer(newViewer(results, testRun, "", scs)))
	defer srv.Close()

	var runs Runs
	getJSON(t, srv.URL+"/api/runs", &runs)
	want := []GameItem{{"a", rec.Status}, {"b", stateLive}}
	if runs.Run != testRun || len(runs.Runs) != 1 || !slices.Equal(runs.Runs[0].Games, want) {
		t.Fatalf("runs %+v, want one run with %v", runs, want)
	}

	snap := finished(t, srv.URL+"/api/steps?run="+testRun+"&game=a")
	if snap.Game != "a" || snap.End == nil || snap.Total != rec.Decisions+1 {
		t.Fatalf("game a: got game %q, %d steps, end %+v", snap.Game, snap.Total, snap.End)
	}

	for _, q := range []string{"run=" + testRun + "&game=../../outside", "run=..&game=outside", "run=" + testRun + "&game=c", "game=a"} {
		if code := getJSON(t, srv.URL+"/api/steps?"+q, &snap); code != http.StatusNotFound {
			t.Fatalf("%s: got %d, want 404", q, code)
		}
	}
}

// Following a run shows its game being played and moves to the next game once
// that one has ended and the next has started.
func TestServerFollowsRunToNextGame(t *testing.T) {
	scs, path, _ := oneTrace(t)
	data, _ := os.ReadFile(path)
	results := runDir(t, map[string][]byte{"a": data})
	srv := httptest.NewServer(newServer(newViewer(results, testRun, "", scs)))
	defer srv.Close()
	url := srv.URL + "/api/steps?run=" + testRun

	if snap := finished(t, url); snap.Game != "a" {
		t.Fatalf("got game %q, want the only game a", snap.Game)
	}
	// b is created and its start entry written, as the next game of the run.
	os.WriteFile(filepath.Join(results, testRun, "live", "b.jsonl"), data[:bytes.IndexByte(data, '\n')+1], 0o644)
	var snap Snapshot
	if getJSON(t, url, &snap); snap.Game != "b" || snap.End != nil {
		t.Fatalf("got game %q (end %+v), want the new live game b", snap.Game, snap.End)
	}
}
