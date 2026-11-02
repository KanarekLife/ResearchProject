package main

import (
	"flag"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
)

// A provider is the endpoint of a run, not the environment's server, and a
// setting it leaves out keeps the default.
func TestProviderSetsEndpoint(t *testing.T) {
	t.Setenv(defaultServerEnv, "http://elsewhere/v1")
	cfg := defaultConfig()
	cfg.Providers = map[string]providerConfig{"proxy": {
		API: apiAnthropic, Server: "http://proxy/v1", APIKeyEnv: "PROXY_KEY", Headers: map[string]string{"X-Run": "{run}"},
	}}
	cfg, err := cfg.withProvider("proxy")
	if err != nil {
		t.Fatal(err)
	}
	_, o := parseRunFlags(flag.NewFlagSet(cmdNameRun, flag.ContinueOnError), nil, cfg)
	if o.api != apiAnthropic || o.server != "http://proxy/v1" || o.headers["X-Run"] != "{run}" || cfg.Inference.APIKeyEnv != "PROXY_KEY" {
		t.Errorf("endpoint %s %s %v %s, want the provider's", o.api, o.server, o.headers, cfg.Inference.APIKeyEnv)
	}
	if o.retries != defaultRetries {
		t.Errorf("retries %d, want the default", o.retries)
	}
	if _, err := cfg.withProvider("nope"); err == nil {
		t.Error("an unknown provider was accepted")
	}
}

// A duration is written as "6h" in config.yaml and in MCBENCH_* variables;
// the worker sets a provider's wait only through the environment.
func TestRateLimitWaitReadsDurations(t *testing.T) {
	var cfg config
	if err := yaml.Unmarshal([]byte("inference: {rate_limit_wait: 6h}"), &cfg); err != nil || cfg.Inference.RateLimitWait != 6*time.Hour {
		t.Fatalf("yaml: %v, err %v", cfg.Inference.RateLimitWait, err)
	}
	if err := applyEnv(&cfg, []string{"MCBENCH_PROVIDERS_GO_RATE_LIMIT_WAIT=90m"}); err != nil {
		t.Fatal(err)
	}
	if w := cfg.Providers["go"].RateLimitWait; w == nil || *w != 90*time.Minute {
		t.Errorf("env: %v", w)
	}
}
