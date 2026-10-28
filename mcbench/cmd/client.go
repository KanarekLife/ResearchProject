package main

import (
	"fmt"
	"os"
	"strings"

	"mcbench/integrations/inference"
	"mcbench/integrations/inference/anthropic"
	"mcbench/integrations/inference/google"
	openaicompatible "mcbench/integrations/inference/openai_compatible"
	openairesponses "mcbench/integrations/inference/openai_responses"
	"mcbench/integrations/systemone"
)

// newClient builds the inference client for the -api protocol. An empty
// server means the provider's own URL, for the protocols that have one.
func newClient(o runSettings, apiKey string) (inference.Client, error) {
	switch o.api {
	case apiOpenAI:
		return openaicompatible.New(openaicompatible.Config{
			BaseURL: o.server, APIKey: apiKey, Model: o.modelID,
			Temperature: o.temp, MaxTokens: o.maxTokens, Retries: o.retries, RateLimitWait: o.rateLimitWait, Headers: o.headers,
			ReasoningEffort: o.effort, ReasoningBudget: o.budget, Stream: o.stream,
		}), nil
	case apiGoogle:
		return google.New(google.Config{
			BaseURL: orDefault(o.server, google.DefaultBaseURL), APIKey: apiKey, Model: o.modelID,
			Temperature: o.temp, MaxTokens: o.maxTokens, Retries: o.retries, RateLimitWait: o.rateLimitWait, Headers: o.headers,
			ReasoningEffort: o.effort, ReasoningBudget: o.budget, Stream: o.stream,
		}), nil
	case apiResponses:
		return openairesponses.New(openairesponses.Config{
			BaseURL: orDefault(o.server, openairesponses.DefaultBaseURL), APIKey: apiKey, Model: o.modelID,
			Temperature: o.temp, MaxTokens: o.maxTokens, Retries: o.retries, RateLimitWait: o.rateLimitWait, Headers: o.headers, ReasoningEffort: o.effort, Stream: o.stream,
		}), nil
	case apiAnthropic:
		return anthropic.New(anthropic.Config{
			BaseURL: orDefault(o.server, anthropic.DefaultBaseURL), APIKey: apiKey, Model: o.modelID,
			Temperature: o.temp, MaxTokens: o.maxTokens, Retries: o.retries, RateLimitWait: o.rateLimitWait, Headers: o.headers,
			ReasoningEffort: o.effort, ReasoningBudget: o.budget, Stream: o.stream,
		}), nil
	}
	return nil, fmt.Errorf("unknown -api %q: use %s, %s, %s or %s", o.api, apiOpenAI, apiResponses, apiGoogle, apiAnthropic)
}

// systemOneClient builds the System One client. The server is -server or the
// server_env variable; without either it is TypeSafe's own, never the local
// chat default.
func systemOneClient(o runSettings, apiKey string) *systemone.Client {
	server := o.server
	if server == defaultServer {
		server = ""
	}
	return systemone.New(systemone.Config{
		BaseURL: orDefault(server, systemone.DefaultBaseURL), APIKey: apiKey, Model: o.modelID,
		Retries: o.retries, RateLimitWait: o.rateLimitWait, Headers: o.headers,
	})
}

// sentReasoning is the reasoning setting the -api client actually sends, so
// that the player label (modelLabel) names only what the model was given:
// chat completions sends both, Gemini and Messages send only the level when
// both are set, and Responses takes no budget.
func sentReasoning(api, effort string, budget int) (string, int) {
	if api == apiResponses || (api != apiOpenAI && effort != "") {
		return effort, 0
	}
	return effort, budget
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// headerFlag collects repeated -header "Name: value" flags.
type headerFlag map[string]string

func (h headerFlag) String() string { return fmt.Sprint(map[string]string(h)) }

func (h headerFlag) Set(s string) error {
	name, value, ok := strings.Cut(s, ":")
	name = strings.TrimSpace(name)
	if !ok || name == "" {
		return fmt.Errorf("header %q: want \"Name: value\"", s)
	}
	value = strings.TrimSpace(value)
	if err := inference.ValidateHeaderValue(value); err != nil {
		return fmt.Errorf("header %q: %w", name, err)
	}
	h[name] = value
	return nil
}

// resolveServer picks the base URL: -server, then the environment variable
// named by server_env, then the configured server. A protocol with its own
// provider URL keeps that one (empty) while the configured server is still the
// OpenAI-compatible default.
func resolveServer(api, flagServer string, flagSet bool, cfg inferenceConfig) string {
	if flagSet {
		return flagServer
	}
	if v := os.Getenv(cfg.ServerEnv); cfg.ServerEnv != "" && v != "" {
		return v
	}
	if api != apiOpenAI && cfg.Server == defaultServer {
		return ""
	}
	return cfg.Server
}
