package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// config mirrors config.yaml. It is loaded only by cmd; every other package
// receives its settings as plain arguments.
type config struct {
	Inference inferenceConfig    `yaml:"inference"`
	Player    playerConfig       `yaml:"player"`
	Run       runConfig          `yaml:"run"`
	Scoring   map[string]float64 `yaml:"scoring"`
}

type inferenceConfig struct {
	API         string  `yaml:"api"` // openai, google or anthropic
	Server      string  `yaml:"server"`
	APIKeyEnv   string  `yaml:"api_key_env"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
	Retries     int     `yaml:"retries"`
	// ReasoningEffort and ReasoningBudget are sent only when set.
	ReasoningEffort string `yaml:"reasoning_effort"`
	ReasoningBudget int    `yaml:"reasoning_budget"`
}

type playerConfig struct {
	Name            string   `yaml:"name"`
	MaxFailures     int      `yaml:"max_failures"`
	InstructionsDir string   `yaml:"instructions_dir"`
	Instructions    []string `yaml:"instructions"`
}

type runConfig struct {
	Seeds    int    `yaml:"seeds"`
	Samples  int    `yaml:"samples"`
	Parallel int    `yaml:"parallel"`
	Out      string `yaml:"out"`
}

func defaultConfig() config {
	return config{
		Inference: inferenceConfig{
			API: apiOpenAI, Server: defaultServer, APIKeyEnv: defaultAPIKeyEnv, Temperature: defaultTemperature,
			MaxTokens: defaultMaxTokens, Retries: defaultRetries,
			ReasoningEffort: defaultReasoningEffort, ReasoningBudget: defaultReasoningBudget,
		},
		Player: playerConfig{
			Name: playerModel, MaxFailures: defaultModelMaxFailures,
			InstructionsDir: defaultInstructionsDir, Instructions: []string{defaultRulesDoc, defaultContractDoc},
		},
		Run:     runConfig{Seeds: defaultSeeds, Samples: defaultSamples, Parallel: defaultParallel, Out: defaultResultsDir},
		Scoring: map[string]float64{},
	}
}

// loadConfig reads path over the built-in defaults. A missing file is fine.
func loadConfig(path string) (config, error) {
	cfg := defaultConfig()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := yaml.UnmarshalWithOptions(data, &cfg, yaml.Strict()); err != nil {
		return cfg, fmt.Errorf("config %s: %w", path, err)
	}
	return cfg, nil
}
