package claudecode

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"

	"mcbench/integrations/inference"
)

// mcpServer is the name Claude Code knows the tools under; it calls them
// mcp__<server>__<tool>.
const (
	mcpServer     = "mcbench"
	mcpToolPrefix = "mcp__" + mcpServer + "__"
	systemFile    = "system.md"
	// stopGrace is how long a process may take to exit after its input closes.
	stopGrace = 5 * time.Second
	// stderrTail is how much of the process's stderr an error quotes.
	stderrTail = 2000
)

// strippedEnv are credentials that would make Claude Code bill an API key
// instead of the logged-in subscription.
var strippedEnv = []string{"ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN"}

// game is one Claude Code process playing one game.
type game struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	events chan event // stdout, closed when the process exits
	calls  chan *call // tool calls from the MCP server
	srv    *http.Server
	stderr *tailBuffer

	sent    int               // messages of the conversation Claude Code has
	open    []*toolUse        // tool uses returned but not yet called
	waiting map[string]*call  // calls awaiting their result, by tool use id
	results map[string]string // results that arrived before their call
	early   []*call           // calls that arrived before their tool use
	limit   *rateLimit        // the usage limit that rejected the last request
	waited  time.Duration     // time spent waiting out usage limits
}

type toolUse struct {
	id, name string
	args     any
}

func start(cfg Config, dir, session string, resume bool, system string, tools []inference.Tool) (*game, error) {
	g := &game{
		events: make(chan event, 64), calls: make(chan *call, 16), stderr: &tailBuffer{max: stderrTail},
		waiting: map[string]*call{}, results: map[string]string{},
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	g.srv = &http.Server{Handler: &mcpHandler{tools: tools, calls: g.calls}}
	go g.srv.Serve(ln)

	if err := os.WriteFile(filepath.Join(dir, systemFile), []byte(system), 0o644); err != nil {
		g.srv.Close()
		return nil, err
	}
	mcp, _ := json.Marshal(map[string]any{"mcpServers": map[string]any{
		mcpServer: map[string]string{"type": "http", "url": "http://" + ln.Addr().String() + "/mcp"},
	}})
	args := []string{
		"-p", "--input-format", "stream-json", "--output-format", "stream-json", "--verbose", "--include-partial-messages",
		"--model", cfg.Model, "--system-prompt-file", systemFile,
		// Only the game tools: no built-in tools, other MCP servers, settings,
		// skills or commands of the machine it runs on.
		"--tools", "", "--strict-mcp-config", "--mcp-config", string(mcp), "--allowedTools", "mcp__" + mcpServer,
		"--setting-sources", "", "--disable-slash-commands",
	}
	if cfg.Effort != "" {
		args = append(args, "--effort", cfg.Effort)
	}
	if resume {
		args = append(args, "--resume", session)
	} else {
		args = append(args, "--session-id", session)
	}
	g.cmd = exec.Command(cfg.Binary, args...)
	g.cmd.Dir = dir
	g.cmd.Env = withoutEnv(os.Environ(), strippedEnv)
	g.cmd.Stderr = g.stderr
	if g.stdin, err = g.cmd.StdinPipe(); err != nil {
		g.srv.Close()
		return nil, err
	}
	stdout, err := g.cmd.StdoutPipe()
	if err != nil {
		g.srv.Close()
		return nil, err
	}
	if err := g.cmd.Start(); err != nil {
		g.srv.Close()
		return nil, fmt.Errorf("claude code: %w", err)
	}
	go g.read(stdout)
	return g, nil
}

// read decodes stdout into events until the process exits.
func (g *game) read(stdout io.Reader) {
	defer close(g.events)
	sc := bufio.NewScanner(stdout)
	sc.Buffer(nil, 1<<26)
	for sc.Scan() {
		var e event
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			slog.Warn("claude code: unreadable output line", "error", err)
			continue
		}
		g.events <- e
	}
	g.cmd.Wait()
}

// deliver hands Claude Code the messages after the ones it has: results go to
// the tool calls waiting on them, user messages open its next turn.
func (g *game) deliver(msgs []inference.Message) error {
	for _, m := range msgs[g.sent:] {
		switch m.Role {
		case inference.RoleTool:
			if c := g.waiting[m.ToolCallID]; c != nil {
				delete(g.waiting, m.ToolCallID)
				c.out <- m.Content
			} else {
				g.results[m.ToolCallID] = m.Content
			}
		case inference.RoleUser:
			if err := g.send(m.Content); err != nil {
				return err
			}
		}
	}
	g.sent = len(msgs)
	return nil
}

func (g *game) send(text string) error {
	b, _ := json.Marshal(userInput{Type: "user", Message: userMessage{Role: "user", Content: text}})
	if _, err := g.stdin.Write(append(b, '\n')); err != nil {
		return g.failure(fmt.Errorf("writing to claude code: %w", err))
	}
	return nil
}

// next waits for Claude Code's next reply, answering tool calls of the
// previous one as their results are known. A request a usage limit rejects
// ends Claude Code's turn; the game waits for the limit to reset (at most
// wait in all) and then opens a new turn with the resume prompt.
func (g *game) next(ctx context.Context, wait time.Duration) (inference.Reply, error) {
	var r replyBuilder
	for {
		select {
		case <-ctx.Done():
			return inference.Reply{}, ctx.Err()
		case c := <-g.calls:
			g.call(c)
		case e, ok := <-g.events:
			if !ok {
				return inference.Reply{}, g.failure(errors.New("claude code exited"))
			}
			if err := g.event(e, &r); err != nil {
				var limit *rateLimit
				if !errors.As(err, &limit) {
					return inference.Reply{}, err
				}
				if err := g.waitOut(ctx, limit, wait); err != nil {
					return inference.Reply{}, err
				}
				r = replyBuilder{}
				continue
			}
			if r.done() {
				reply := r.reply()
				g.open = r.uses
				for _, c := range g.early {
					g.call(c)
				}
				g.early = nil
				return reply, nil
			}
		}
	}
}

func (g *game) event(e event, r *replyBuilder) error {
	switch e.Type {
	case eventStream:
		r.stream(e.Event)
	case eventAssistant:
		r.assistant(e.Message)
	case eventResult:
		if e.IsError {
			if limit := g.limit; limit != nil {
				g.limit = nil
				return limit
			}
			return g.failure(fmt.Errorf("claude code: %s %s", e.Subtype, e.Result))
		}
	case eventRateLimit:
		if e.RateLimit.Status == rateLimitRejected {
			g.limit = &rateLimit{kind: e.RateLimit.Type, resets: time.Unix(e.RateLimit.ResetsAt, 0)}
		}
	case eventSystem:
		if e.Subtype == subtypeCompact {
			slog.Warn("claude code compacted the conversation; it is no longer append-only")
		}
	}
	return nil
}

// rateLimit is a subscription usage limit that rejected a request.
type rateLimit struct {
	kind   string
	resets time.Time
}

func (l *rateLimit) Error() string {
	return fmt.Sprintf("claude code: %s usage limit reached, resets at %s", l.kind, l.resets.Format(time.RFC3339))
}

// resetMargin is waited past a limit's reset time, so the next request is
// not rejected again by a clock that runs a little behind.
const resetMargin = time.Minute

// waitOut waits until the limit resets and reopens the turn, unless that
// would take the game's waiting past wait.
func (g *game) waitOut(ctx context.Context, limit *rateLimit, wait time.Duration) error {
	d := time.Until(limit.resets) + resetMargin
	if limit.resets.IsZero() || g.waited+d > wait {
		return fmt.Errorf("%w (waited %s of %s)", limit, g.waited.Round(time.Second), wait)
	}
	slog.Warn("claude code usage limit reached, waiting", "limit", limit.kind, "wait", d.Round(time.Second), "resets", limit.resets)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
	}
	g.waited += d
	slog.Warn("claude code usage limit reset; continuing the game with an untraced user message", "message", resumePrompt)
	return g.send(resumePrompt)
}

// call matches a tool call to the tool use it runs and answers it once its
// result is known. Claude Code may start a call before the reply that asks
// for it is complete; such a call waits for the reply.
func (g *game) call(c *call) {
	for i, u := range g.open {
		if c.name == u.name && reflect.DeepEqual(c.args, u.args) {
			g.open = append(g.open[:i], g.open[i+1:]...)
			if out, ok := g.results[u.id]; ok {
				delete(g.results, u.id)
				c.out <- out
			} else {
				g.waiting[u.id] = c
			}
			return
		}
	}
	g.early = append(g.early, c)
}

// failure adds what the process wrote to stderr to err.
func (g *game) failure(err error) error {
	if tail := strings.TrimSpace(g.stderr.String()); tail != "" {
		return fmt.Errorf("%w: %s", err, tail)
	}
	return err
}

func (g *game) stop() {
	g.stdin.Close()
	exited := make(chan struct{})
	go func() {
		for range g.events {
		}
		close(exited)
	}()
	select {
	case <-exited:
	case <-time.After(stopGrace):
		g.cmd.Process.Kill()
		<-exited
	}
	g.srv.Close()
}

func withoutEnv(env, names []string) []string {
	out := env[:0:0]
	for _, kv := range env {
		name, _, _ := strings.Cut(kv, "=")
		if !slices.Contains(names, name) {
			out = append(out, kv)
		}
	}
	return out
}

// tailBuffer keeps the last max bytes written to it.
type tailBuffer struct {
	mu  sync.Mutex
	max int
	b   []byte
}

func (t *tailBuffer) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.b = append(t.b, p...)
	if len(t.b) > t.max {
		t.b = t.b[len(t.b)-t.max:]
	}
	return len(p), nil
}

func (t *tailBuffer) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return string(t.b)
}
