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
	Data      dataConfig         `yaml:"data"`
	Inference inferenceConfig    `yaml:"inference"`
	Player    playerConfig       `yaml:"player"`
	Run       runConfig          `yaml:"run"`
	Scoring   map[string]float64 `yaml:"scoring"`
}

type dataConfig struct {
	Scenarios string `yaml:"scenarios"`
	Root      string `yaml:"root"`
}

type inferenceConfig struct {
	Server      string  `yaml:"server"`
	APIKeyEnv   string  `yaml:"api_key_env"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
	Retries     int     `yaml:"retries"`
}

type playerConfig struct {
	Name            string   `yaml:"name"`
	MaxSteps        int      `yaml:"max_steps"`
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
		Data: dataConfig{Scenarios: defaultScenarioDir, Root: defaultDataRoot},
		Inference: inferenceConfig{
			Server: defaultServer, APIKeyEnv: defaultAPIKeyEnv,
			MaxTokens: defaultMaxTokens, Retries: defaultRetries,
		},
		Player: playerConfig{
			Name: playerModel, MaxSteps: defaultModelMaxSteps,
			InstructionsDir: defaultInstructionsDir, Instructions: []string{defaultInstruction},
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
