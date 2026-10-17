package main

import (
	"fmt"

	"mcbench/integrations/inference"
	"mcbench/integrations/inference/anthropic"
	"mcbench/integrations/inference/google"
	openaicompatible "mcbench/integrations/inference/openai_compatible"
	openairesponses "mcbench/integrations/inference/openai_responses"
)

// newClient builds the inference client for the -api protocol. An empty
// server means the provider's own URL, for the protocols that have one.
func newClient(o runSettings, apiKey string) (inference.Client, error) {
	switch o.api {
	case apiOpenAI:
		return openaicompatible.New(openaicompatible.Config{
			BaseURL: o.server, APIKey: apiKey, Model: o.modelID,
			Temperature: o.temp, MaxTokens: o.maxTokens, Retries: o.retries,
			ReasoningEffort: o.effort, ReasoningBudget: o.budget,
		}), nil
	case apiGoogle:
		return google.New(google.Config{
			BaseURL: orDefault(o.server, google.DefaultBaseURL), APIKey: apiKey, Model: o.modelID,
			Temperature: o.temp, MaxTokens: o.maxTokens, Retries: o.retries,
			ReasoningEffort: o.effort, ReasoningBudget: o.budget,
		}), nil
	case apiResponses:
		return openairesponses.New(openairesponses.Config{
			BaseURL: orDefault(o.server, openairesponses.DefaultBaseURL), APIKey: apiKey, Model: o.modelID,
			Temperature: o.temp, MaxTokens: o.maxTokens, Retries: o.retries, ReasoningEffort: o.effort,
		}), nil
	case apiAnthropic:
		return anthropic.New(anthropic.Config{
			BaseURL: orDefault(o.server, anthropic.DefaultBaseURL), APIKey: apiKey, Model: o.modelID,
			Temperature: o.temp, MaxTokens: o.maxTokens, Retries: o.retries,
			ReasoningEffort: o.effort, ReasoningBudget: o.budget,
		}), nil
	}
	return nil, fmt.Errorf("unknown -api %q: use %s, %s, %s or %s", o.api, apiOpenAI, apiResponses, apiGoogle, apiAnthropic)
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
