package main

// Command names accepted on the command line.
const (
	cmdNameList     = "list"
	cmdNameValidate = "validate"
	cmdNamePlay     = "play"
	cmdNameRun      = "run"
	cmdNameReport   = "report"
	cmdNameAnalyze  = "analyze"
	cmdNameView     = "view"
	cmdNameBoardGUI = "board_gui"
	cmdNameSite     = "site"
	cmdNamePublish  = "publish"
	cmdNameJobs     = "jobs"
)

// Player names accepted by -player.
const (
	playerModel     = "model"
	playerSystemOne = "systemone"
	playerHeuristic = "heuristic"
	playerRandom    = "random"
	playerFirst     = "first"
)

// API protocols accepted by -api.
const (
	apiOpenAI    = "openai"
	apiGoogle    = "google"
	apiAnthropic = "anthropic"
	apiResponses = "responses"
)

// defaultRecentEvents is how many log lines the systemone player sends.
const defaultRecentEvents = 10

// randomPlayerSeed makes the random baseline reproducible.
const randomPlayerSeed = 1

// Default configuration values, overridden by config.yaml and flags.
const (
	defaultConfigFile      = "config.yaml"
	defaultInstructionsDir = "data/instructions"
	defaultResultsDir      = "results"
	defaultServer          = "http://localhost:1234/v1"
	defaultAPIKeyEnv       = "OPENAI_API_KEY"
	defaultServerEnv       = "OPENAI_BASE_URL"
	defaultRulesDoc        = "rules_agent"
	defaultContractDoc     = "mcbench"
	// defaultTemperature is not sent: the provider's default applies.
	defaultTemperature      = -1
	defaultModelMaxFailures = 5
	defaultMaxTokens        = 32768
	defaultRetries          = 3
	// Reasoning defaults: the level and the thinking token cap.
	defaultReasoningEffort = "medium"
	defaultReasoningBudget = 2048
	defaultSeeds           = 10
	defaultSamples         = 1
	defaultParallel        = 1
	defaultRemote          = "origin"
	defaultResultsBranch   = "results"
	defaultOnlineDir       = "online-results"
	defaultJobsBranch      = "jobs"
	defaultJobsDir         = "online-jobs"
	defaultPublish         = true
)

// Result sources accepted by -source.
const (
	sourceLocal  = "local"
	sourceOnline = "online"
	sourceHybrid = "hybrid"
)

// Subdirectories written under the results directory.
const (
	liveSubdir  = "live"
	gamesFile   = "games.jsonl"
	runIDLayout = "20060102T150405Z"
	runNameSep  = "__" // between the player and the instruction documents in a run directory name
)
