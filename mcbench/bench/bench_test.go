package bench

import (
	"context"
	"reflect"
	"testing"

	"mcbench/agent"
	"mcbench/game"
	"mcbench/scenario"
)

func scenarios(t *testing.T) []*scenario.Scenario {
	t.Helper()
	scs, err := scenario.LoadDir("../scenarios", "..")
	if err != nil {
		t.Fatal(err)
	}
	return scs
}

// The same scenario, seed and choices must give the same game.
func TestDeterministic(t *testing.T) {
	for _, sc := range scenarios(t) {
		a := PlayGame(context.Background(), sc, agent.Random{Seed: 42}, agent.Instructions{}, sc.Seeds[0], 0)
		b := PlayGame(context.Background(), sc, agent.Random{Seed: 42}, agent.Instructions{}, sc.Seeds[0], 0)
		a.Millis, b.Millis = 0, 0
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("%s: two identical games differ", sc.ID)
		}
	}
}

// Every seed must play to the end without engine errors.
func TestGamesComplete(t *testing.T) {
	for _, sc := range scenarios(t) {
		for _, seed := range sc.Seeds {
			for _, ag := range []agent.Agent{agent.Random{Seed: seed}, agent.Heuristic{}} {
				r := PlayGame(context.Background(), sc, ag, agent.Instructions{}, seed, 0)
				if r.Status == game.EngineError || r.Status == game.DecisionLimit || r.Status == "agent_error" {
					t.Fatalf("%s seed %d (%s): %s %s", sc.ID, seed, ag.Name(), r.Status, r.Error)
				}
			}
		}
	}
}
