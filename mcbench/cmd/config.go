package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

// config mirrors config.yaml. It is loaded only by cmd; every other package
// receives its settings as plain arguments.
type config struct {
	Inference inferenceConfig    `yaml:"inference"`
	Player    playerConfig       `yaml:"player"`
	Run       runConfig          `yaml:"run"`
	Online    onlineConfig       `yaml:"online"`
	Jobs      jobsConfig         `yaml:"jobs"`
	Scoring   map[string]float64 `yaml:"scoring"`
	// Providers are the endpoints a job or `run -provider` names, kept on the
	// worker so that keys and headers never reach a jobs branch.
	Providers map[string]providerConfig `yaml:"providers"`
}

// providerConfig is how to reach a model, not what is measured. A field left
// out keeps the inference default.
type providerConfig struct {
	API           string            `yaml:"api"`
	Server        string            `yaml:"server"` // empty: the protocol's own URL
	APIKeyEnv     string            `yaml:"api_key_env"`
	Headers       map[string]string `yaml:"headers"`
	Stream        *bool             `yaml:"stream"`
	Retries       *int              `yaml:"retries"`
	RateLimitWait *time.Duration    `yaml:"rate_limit_wait"`
	// Usage names the plan whose usage windows the jobs worker reads before it
	// takes a job for this provider: opencode or commandcode; empty never waits.
	Usage string `yaml:"usage"`
}

// jobsConfig is the branch of job files on online.remote and its checkout.
type jobsConfig struct {
	Branch string `yaml:"branch"`
	Dir    string `yaml:"dir"`
	// MaxUsage is the percent of a provider's usage window at which the
	// worker leaves that provider's jobs pending until the window resets.
	MaxUsage float64 `yaml:"max_usage"`
}

type inferenceConfig struct {
	API       string `yaml:"api"` // openai, google or anthropic
	Server    string `yaml:"server"`
	APIKeyEnv string `yaml:"api_key_env"`
	// ServerEnv names the variable that overrides Server.
	ServerEnv   string  `yaml:"server_env"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
	Retries     int     `yaml:"retries"`
	// RateLimitWait is how long rate limits are waited out per request.
	RateLimitWait time.Duration `yaml:"rate_limit_wait"`
	// ReasoningEffort and ReasoningBudget are sent only when set.
	ReasoningEffort string `yaml:"reasoning_effort"`
	ReasoningBudget int    `yaml:"reasoning_budget"`
	// Stream asks for streamed replies, whichever protocol is used.
	Stream bool `yaml:"stream"`
	// Headers come only from a provider.
	Headers map[string]string `yaml:"-"`
}

type playerConfig struct {
	Name            string   `yaml:"name"`
	MaxFailures     int      `yaml:"max_failures"`
	InstructionsDir string   `yaml:"instructions_dir"`
	Instructions    []string `yaml:"instructions"`
}

type runConfig struct {
	// Scenarios are the scenario ids run and play use; empty means all.
	Scenarios []string `yaml:"scenarios"`
	Seeds     int      `yaml:"seeds"`
	Samples   int      `yaml:"samples"`
	Parallel  int      `yaml:"parallel"`
	Out       string   `yaml:"out"`
}

type onlineConfig struct {
	// Repo is the repository the branches are pushed from; empty: the current
	// directory's. The worker image has none of its own.
	Repo    string `yaml:"repo"`
	Remote  string `yaml:"remote"`
	Branch  string `yaml:"branch"`
	Dir     string `yaml:"dir"`
	Publish bool   `yaml:"publish"`
	Source  string `yaml:"source"`
}

func defaultConfig() config {
	return config{
		Inference: inferenceConfig{
			API: apiOpenAI, Server: defaultServer, APIKeyEnv: defaultAPIKeyEnv, ServerEnv: defaultServerEnv, Temperature: defaultTemperature,
			MaxTokens: defaultMaxTokens, Retries: defaultRetries, RateLimitWait: defaultRateLimitWait,
			ReasoningEffort: defaultReasoningEffort, ReasoningBudget: defaultReasoningBudget,
		},
		Player: playerConfig{
			Name: playerModel, MaxFailures: defaultModelMaxFailures,
			InstructionsDir: defaultInstructionsDir, Instructions: []string{defaultRulesDoc, defaultContractDoc},
		},
		Run: runConfig{Seeds: defaultSeeds, Samples: defaultSamples, Parallel: defaultParallel, Out: defaultResultsDir},
		Online: onlineConfig{
			Remote: defaultRemote, Branch: defaultResultsBranch, Dir: defaultOnlineDir,
			Publish: defaultPublish, Source: sourceHybrid,
		},
		Jobs:    jobsConfig{Branch: defaultJobsBranch, Dir: defaultJobsDir, MaxUsage: defaultMaxUsage},
		Scoring: map[string]float64{},
	}
}

// loadConfig reads path over the built-in defaults, then the MCBENCH_*
// variables over both. A missing file is fine.
func loadConfig(path string) (config, error) {
	cfg := defaultConfig()
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return cfg, err
	}
	if err == nil {
		if err := yaml.UnmarshalWithOptions(data, &cfg, yaml.Strict()); err != nil {
			return cfg, fmt.Errorf("config %s: %w", path, err)
		}
	}
	return cfg, applyEnv(&cfg, os.Environ())
}

// withProvider makes a provider's endpoint settings the inference defaults.
// Its server is used as given: server_env no longer overrides it.
func (cfg config) withProvider(name string) (config, error) {
	p, ok := cfg.Providers[name]
	if !ok {
		return cfg, fmt.Errorf("provider %q is not in the config's providers", name)
	}
	in := &cfg.Inference
	in.Server, in.ServerEnv = p.Server, ""
	if p.API != "" {
		in.API = p.API
	}
	if p.APIKeyEnv != "" {
		in.APIKeyEnv = p.APIKeyEnv
	}
	if p.Stream != nil {
		in.Stream = *p.Stream
	}
	if p.Retries != nil {
		in.Retries = *p.Retries
	}
	if p.RateLimitWait != nil {
		in.RateLimitWait = *p.RateLimitWait
	}
	in.Headers = headerFlag{}
	for k, v := range p.Headers {
		if err := headerFlag(in.Headers).Set(k + ": " + v); err != nil {
			return cfg, fmt.Errorf("provider %q: %w", name, err)
		}
	}
	return cfg, nil
}
