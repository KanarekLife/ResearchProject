package benchmark

import (
	"strings"

	"mcbench/constants"
)

// GameStats is one game as the statistics page reads it.
type GameStats struct {
	Run          string             `json:"run"`  // the run directory, as the board names it
	Game         string             `json:"game"` // the trace name in the run
	Scenario     string             `json:"scenario"`
	Player       string             `json:"player"`
	Instructions string             `json:"instructions"` // names joined by ",", empty for none
	InstrSHA     string             `json:"instructions_sha,omitempty"`
	Seed         uint64             `json:"seed"`
	Sample       int                `json:"sample"`
	Status       string             `json:"status"`
	Error        string             `json:"error,omitempty"`
	Score        float64            `json:"score"`
	Criteria     map[string]float64 `json:"criteria"`
	Rounds       int                `json:"rounds"`
	Decisions    int                `json:"decisions"`
	Invalid      int                `json:"invalid"`
	Stats        map[string]float64 `json:"stats"` // the final board facts the criteria come from
	Requests     int                `json:"requests,omitempty"`
	TokensIn     int64              `json:"tokens_in"`
	TokensOut    int64              `json:"tokens_out"`
	TokensCached int64              `json:"tokens_cached,omitempty"`
	Millis       int64              `json:"ms"`
	// Behaviour is empty for an infra_error game, which Analyze leaves out.
	Behaviour Behaviour `json:"behaviour"`
}

// StatsOf is a record's GameStats in run.
func StatsOf(run string, r Record) GameStats {
	g := GameStats{
		Run: run, Game: gameName(r.Scenario, r.Seed, r.Sample), Scenario: r.Scenario, Player: r.Player,
		Instructions: strings.Join(r.Instructions.Names, ","), InstrSHA: r.Instructions.SHA256,
		Seed: r.Seed, Sample: r.Sample, Status: r.Status, Error: firstLine(r.Error),
		Score: r.Score, Criteria: r.Criteria, Rounds: r.Rounds, Decisions: r.Decisions, Invalid: r.Invalid,
		Stats: r.Stats, TokensIn: r.Usage.InputTokens, TokensOut: r.Usage.OutputTokens,
		TokensCached: r.Usage.CachedTokens, Requests: r.Usage.Requests, Millis: r.Millis,
	}
	if r.Status != constants.InfraError {
		g.Behaviour.add(r)
	}
	return g
}
