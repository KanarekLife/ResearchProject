package main

import (
	"flag"
	"testing"
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
