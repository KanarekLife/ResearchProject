package main

import (
	"reflect"
	"testing"
)

func TestEnvOverridesConfig(t *testing.T) {
	cfg := defaultConfig()
	cfg.Providers = map[string]providerConfig{"proxy": {API: apiOpenAI}}
	err := applyEnv(&cfg, []string{
		"HOME=/root",
		"MCBENCH_RUN_OUT=/data/results",
		"MCBENCH_RUN_SEEDS=3",
		"MCBENCH_PLAYER_INSTRUCTIONS=[rules_short]",
		"MCBENCH_INFERENCE_MODEL=0123", // a string stays as written
		"MCBENCH_PROVIDERS_PROXY_API_KEY_ENV=PROXY_KEY",
		"MCBENCH_PROVIDERS_MY_GATEWAY_SERVER=http://gw:8317/v1",
		"MCBENCH_PROVIDERS_MY_GATEWAY_STREAM=true",
		`MCBENCH_PROVIDERS_MY_GATEWAY_HEADERS={X-Run: "{run}"}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Run.Out != "/data/results" || cfg.Run.Seeds != 3 || cfg.Inference.Model != "0123" ||
		!reflect.DeepEqual(cfg.Player.Instructions, []string{"rules_short"}) {
		t.Errorf("run %+v model %q instructions %v", cfg.Run, cfg.Inference.Model, cfg.Player.Instructions)
	}
	if p := cfg.Providers["proxy"]; p.API != apiOpenAI || p.APIKeyEnv != "PROXY_KEY" {
		t.Errorf("proxy %+v: want its file api kept and the env key variable", p)
	}
	gw := cfg.Providers["my_gateway"]
	if gw.Server != "http://gw:8317/v1" || gw.Stream == nil || !*gw.Stream || gw.Headers["X-Run"] != "{run}" {
		t.Errorf("my_gateway %+v", gw)
	}
	for _, bad := range []string{"MCBENCH_RUN_OUTT=x", "MCBENCH_SCORING_WIN=1", "MCBENCH_RUN_SEEDS=many"} {
		if err := applyEnv(&cfg, []string{bad}); err == nil {
			t.Errorf("%s accepted", bad)
		}
	}
}
