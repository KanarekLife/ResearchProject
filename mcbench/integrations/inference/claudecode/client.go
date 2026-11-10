// Package claudecode reaches Claude through the Claude Code CLI (`claude -p`),
// so a run is billed to the CLI's own login (a claude.ai subscription), never
// to an API key. Claude Code runs the agentic loop; the client serves the
// caller's tools to it over a local MCP server and hands each call back to the
// caller as a tool call, so the caller still executes, counts and traces it.
package claudecode

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"mcbench/integrations/inference"
)

// DefaultBinary is the Claude Code executable looked up on PATH.
const DefaultBinary = "claude"

// Config is how to start Claude Code for each game.
type Config struct {
	Binary   string // the claude executable
	Model    string // --model
	Effort   string // --effort, when set
	StateDir string // holds each game's working directory and session id
	// RateLimitWait is how long a game may wait in all for usage limits to
	// reset; 0 ends the game at the first limit.
	RateLimitWait time.Duration
}

// Client plays each game in its own Claude Code process, started on the
// game's first request and stopped by EndGame.
type Client struct {
	cfg   Config
	mu    sync.Mutex
	games map[string]*game
}

func New(cfg Config) *Client {
	if cfg.Binary == "" {
		cfg.Binary = DefaultBinary
	}
	return &Client{cfg: cfg, games: map[string]*game{}}
}

// Complete hands Claude Code the messages it has not seen (tool results and
// user turns; its own replies it already has) and returns its next reply.
func (c *Client) Complete(ctx context.Context, msgs []inference.Message, tools []inference.Tool) (inference.Reply, error) {
	g, err := c.game(ctx, msgs, tools)
	if err != nil {
		return inference.Reply{}, err
	}
	if err := g.deliver(msgs); err != nil {
		return inference.Reply{}, err
	}
	return g.next(ctx, c.cfg.RateLimitWait)
}

// EndGame stops the game's Claude Code process.
func (c *Client) EndGame(ctx context.Context) {
	key := gameKey(ctx)
	c.mu.Lock()
	g := c.games[key]
	delete(c.games, key)
	c.mu.Unlock()
	if g != nil {
		g.stop()
	}
}

// game returns the running game, or starts one: a new conversation, or a
// resumed one when msgs already holds replies the process has not produced.
func (c *Client) game(ctx context.Context, msgs []inference.Message, tools []inference.Tool) (*game, error) {
	key := gameKey(ctx)
	c.mu.Lock()
	defer c.mu.Unlock()
	if g := c.games[key]; g != nil {
		return g, nil
	}
	if len(msgs) == 0 || msgs[0].Role != inference.RoleSystem {
		return nil, errors.New("claude code: the conversation must open with the system prompt")
	}
	dir := filepath.Join(c.cfg.StateDir, filepath.FromSlash(key))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	resume := hasReply(msgs)
	session, err := sessionID(dir, resume)
	if err != nil {
		return nil, err
	}
	g, err := start(c.cfg, dir, session, resume, msgs[0].Content, tools)
	if err != nil {
		return nil, err
	}
	if resume {
		// Claude Code continues its own record of the conversation; a turn
		// opens only on a user message, so one is sent that is not traced.
		g.sent = len(msgs)
		if err := g.send(resumePrompt); err != nil {
			g.stop()
			return nil, err
		}
	} else {
		g.sent = 1
	}
	c.games[key] = g
	return g, nil
}

// resumePrompt reopens a resumed conversation.
const resumePrompt = "The game was interrupted and has been resumed. Continue playing."

const sessionFile = "session"

// sessionID is the game's Claude Code session: a new one, written down so a
// resumed game can continue it, or the one written down before.
func sessionID(dir string, resume bool) (string, error) {
	path := filepath.Join(dir, sessionFile)
	if resume {
		b, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("claude code: no session to resume: %w", err)
		}
		return strings.TrimSpace(string(b)), nil
	}
	// Claude Code wants a UUID: a version 4 one from random bytes.
	u := make([]byte, 16)
	rand.Read(u)
	u[6] = u[6]&0x0f | 0x40
	u[8] = u[8]&0x3f | 0x80
	id := fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
	return id, os.WriteFile(path, []byte(id), 0o644)
}

func hasReply(msgs []inference.Message) bool {
	for _, m := range msgs {
		if m.Role == inference.RoleAssistant {
			return true
		}
	}
	return false
}

// gameKey names the game a request belongs to: its run and game, which are
// unique per game and stable when a run is resumed.
func gameKey(ctx context.Context) string {
	run, name := inference.GameFrom(ctx)
	return run + "/" + name
}
