package main

// Command names accepted on the command line.
const (
	cmdNameList     = "list"
	cmdNameTools    = "tools"
	cmdNameShow     = "show"
	cmdNameValidate = "validate"
	cmdNamePlay     = "play"
	cmdNameRun      = "run"
	cmdNameReport   = "report"
)

// Player names accepted by -player.
const (
	playerModel     = "model"
	playerHeuristic = "heuristic"
	playerRandom    = "random"
	playerFirst     = "first"
)

// Default configuration values, overridden by config.yaml and flags.
const (
	defaultConfigFile      = "config.yaml"
	defaultScenarioDir     = "data/scenarios"
	defaultDataRoot        = "data"
	defaultInstructionsDir = "data/instructions"
	defaultResultsDir      = "results"
	defaultServer          = "http://localhost:1234/v1"
	defaultAPIKeyEnv       = "OPENAI_API_KEY"
	defaultInstruction     = "rules"
	defaultModelMaxSteps   = 5
	defaultMaxTokens       = 32768
	defaultRetries         = 3
	defaultSeeds           = 10
	defaultSamples         = 1
	defaultParallel        = 1
)

// Subdirectories written under the results directory.
const (
	liveSubdir  = "live"
	gamesFile   = "games.jsonl"
	runIDLayout = "20060102T150405Z"
)
