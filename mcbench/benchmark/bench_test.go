package benchmark

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"testing"

	"mcbench/constants"
	"mcbench/game/scenario"
	"mcbench/game/session"
	"mcbench/player"
	"mcbench/player/heuristic"
	"mcbench/player/instruction"
)

func scenarios(t *testing.T) []*scenario.Scenario {
	t.Helper()
	scs, err := scenario.LoadDir("../data/scenarios", "../data")
	if err != nil {
		t.Fatal(err)
	}
	return scs
}

// The same scenario, seed and choices must give the same game.
func TestDeterministic(t *testing.T) {
	for _, sc := range scenarios(t) {
		a := Play(context.Background(), sc, heuristic.Random{Seed: 42}, instruction.Instructions{}, sc.Seeds[0], 0, "", DefaultWeights)
		b := Play(context.Background(), sc, heuristic.Random{Seed: 42}, instruction.Instructions{}, sc.Seeds[0], 0, "", DefaultWeights)
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
			for _, p := range []player.Player{heuristic.Random{Seed: seed}, heuristic.Heuristic{}} {
				r := Play(context.Background(), sc, p, instruction.Instructions{}, seed, 0, "", DefaultWeights)
				if r.Status == session.EngineError || r.Status == session.DecisionLimit || r.Status == constants.AgentError {
					t.Fatalf("%s seed %d (%s): %s %s", sc.ID, seed, p.Name(), r.Status, r.Error)
				}
			}
		}
	}
}

// Games log every action at info level; keep test output readable.
func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	os.Exit(m.Run())
}

// A game logs its actions and choices at info level, tagged with the game.
func TestGameLogsActionsAndChoices(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(prev)

	sc := scenarios(t)[0]
	Play(context.Background(), sc, heuristic.Random{Seed: 1}, instruction.Instructions{}, sc.Seeds[0], 0, "", DefaultWeights)
	out := buf.String()
	for _, want := range []string{"msg=choice", "scenario=" + sc.ID, "round=1", "Player turn"} {
		if !strings.Contains(out, want) {
			t.Errorf("log lacks %q", want)
		}
	}
}
