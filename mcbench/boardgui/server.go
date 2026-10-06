// Package boardgui shows a game's trace as a board in the browser, one decision
// at a time, and follows a game that is still being played. It only reads.
package boardgui

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"

	"mcbench/game/cards"
	"mcbench/game/scenario"
	"mcbench/trace"
)

// poll is how long one read of a growing trace waits for new lines.
const poll = 500 * time.Millisecond

// Serve serves the games under the results directory at addr until the server
// fails, opening run and game first. An empty game follows the run: it shows
// the game being played and moves on to the next one when it ends.
func Serve(addr, results, run, game string, scs []*scenario.Scenario) error {
	return http.ListenAndServe(addr, newServer(newViewer(results, run, game, scs)))
}

// maxReplays bounds the replays kept, so switching back to a game is instant
// and a live game keeps being followed.
const maxReplays = 8

// viewer serves the traces found under one results directory. A trace is
// chosen by run id and game name from that listing, never by path.
type viewer struct {
	results, run, game string
	scs                []*scenario.Scenario

	mu      sync.Mutex
	replays map[string]*replay
	recent  []string // trace paths, least recently used first
}

func newViewer(results, run, game string, scs []*scenario.Scenario) *viewer {
	return &viewer{results: results, run: run, game: game, scs: scs, replays: map[string]*replay{}}
}

// pick resolves a run id and game name, or the run's current game when game is
// empty, to a discovered game.
func (v *viewer) pick(runID, game string) (trace.Game, bool) {
	runs, err := trace.Discover(v.results)
	if err != nil {
		return trace.Game{}, false
	}
	i := slices.IndexFunc(runs, func(r trace.Run) bool { return r.ID == runID })
	if i < 0 {
		return trace.Game{}, false
	}
	if game == "" {
		return current(runs[i]), true
	}
	j := slices.IndexFunc(runs[i].Games, func(g trace.Game) bool { return g.Name == game })
	if j < 0 {
		return trace.Game{}, false
	}
	return runs[i].Games[j], true
}

// replay returns the replay of a trace, starting it on first use.
func (v *viewer) replay(path string) *replay {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.recent = slices.DeleteFunc(v.recent, func(p string) bool { return p == path })
	v.recent = append(v.recent, path)
	r, ok := v.replays[path]
	if !ok {
		r = newReplay(v.scs)
		v.replays[path] = r
		go r.follow(path, poll)
	}
	if len(v.recent) > maxReplays {
		old := v.recent[0]
		v.recent = v.recent[1:]
		v.replays[old].stop()
		delete(v.replays, old)
	}
	return r
}

// The page is a React app in web/, built into static/.
//go:generate npm --prefix web ci
//go:generate npm --prefix web run build

//go:embed static
var static embed.FS

// notBuilt is the placeholder served until the page is built.
const notBuilt = "not-built.txt"

// newServer serves the page, the runs and the replays. Every route is GET:
// nothing the viewer serves can change a trace or a game.
func newServer(v *viewer) http.Handler {
	page, err := fs.Sub(static, "static")
	if err != nil {
		panic(err) // the embedded directory is always there
	}
	files := http.FileServerFS(page)
	if _, err := fs.Stat(page, "index.html"); err != nil {
		files = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { http.ServeFileFS(w, req, page, notBuilt) })
	}
	codes := cardCodes()
	mux := http.NewServeMux()
	mux.Handle("GET /", files)
	mux.HandleFunc("GET /api/cards", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, codes) })
	mux.HandleFunc("GET /api/runs", func(w http.ResponseWriter, _ *http.Request) {
		runs, err := listRuns(v.results)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, Runs{Runs: runs, Run: v.run, Game: v.game})
	})
	mux.HandleFunc("GET /api/steps", func(w http.ResponseWriter, req *http.Request) {
		q := req.URL.Query()
		g, ok := v.pick(q.Get("run"), q.Get("game"))
		if !ok {
			http.Error(w, "no such run or game under the results directory", http.StatusNotFound)
			return
		}
		from, _ := strconv.Atoi(q.Get("from"))
		snap := v.replay(g.Trace).snapshot(from)
		snap.Run, snap.Game = q.Get("run"), g.Name
		writeJSON(w, snap)
	})
	return mux
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(v)
}

// cardCodes maps each card name to its codes in code order. Our card codes are
// MarvelCDB's, so the page can show a card's image from marvelcdb.com. Only
// villain stages share a name; the page picks the code by stage.
func cardCodes() map[string][]string {
	out := map[string][]string{}
	for _, d := range cards.All() {
		out[d.Name] = append(out[d.Name], d.Code)
	}
	return out
}
