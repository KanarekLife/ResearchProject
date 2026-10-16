package benchmark

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"testing"

	"mcbench/constants"
	"mcbench/data"
	"mcbench/game/scenario"
	"mcbench/game/session"
	"mcbench/player"
	"mcbench/player/heuristic"
	"mcbench/player/instruction"
)

func scenarios(t *testing.T) []*scenario.Scenario {
	t.Helper()
	scs, err := scenario.LoadAll(data.FS)
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
				if r.Status == constants.EngineError || r.Status == constants.DecisionLimit || r.Status == constants.AgentError {
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

// failing is a player that stops on its first decision with err.
type failing struct{ err error }

func (failing) Name() string { return "failing" }
func (f failing) Play(context.Context, *session.Session) (player.Usage, error) {
	return player.Usage{}, f.err
}

// Only a failure of the model itself is graded (as 0); a game the model could
// not be reached for is recorded but left out of the report's means.
func TestInfraErrorsAreNotGraded(t *testing.T) {
	sc := scenarios(t)[0]
	play := func(err error) Record {
		return Play(context.Background(), sc, failing{err}, instruction.Instructions{}, sc.Seeds[0], 0, "", DefaultWeights)
	}
	infra := play(fmt.Errorf("%w: chat API HTTP 503", player.ErrInference))
	agent := play(errors.New("no valid decision after 5 model requests"))
	if infra.Status != constants.InfraError || agent.Status != constants.AgentError {
		t.Fatalf("statuses %s and %s", infra.Status, agent.Status)
	}
	s := summarize([]Record{infra, agent})
	if len(s.score.xs) != 1 || s.status[constants.InfraError] != 1 {
		t.Errorf("graded %d games, statuses %v", len(s.score.xs), s.status)
	}
}

// Replaying a game's recorded choices on a fresh session must reach the same
// result: the choices alone determine the game.
func TestReplayReachesSameResult(t *testing.T) {
	for _, sc := range scenarios(t) {
		seed := sc.Seeds[0]
		rec := Play(context.Background(), sc, heuristic.Random{Seed: 7}, instruction.Instructions{}, seed, 0, "", DefaultWeights)
		s, err := session.New(sc, seed, session.Options{MaxRounds: sc.MaxRounds, MaxDecisions: sc.MaxDecisions})
		if err != nil {
			t.Fatal(err)
		}
		for i, c := range rec.Choices {
			id := 0
			for _, o := range s.Decision().Options {
				if o.Key == c.Key {
					id = o.ID
					break
				}
			}
			if id == 0 {
				t.Fatalf("%s: choice %d (%s) is not offered on replay", sc.ID, i+1, c.Key)
			}
			if _, err := s.Choose(id, ""); err != nil {
				t.Fatalf("%s: choice %d: %v", sc.ID, i+1, err)
			}
		}
		if s.Status() != rec.Status || !reflect.DeepEqual(s.Stats(), rec.Stats) {
			t.Errorf("%s: replay ended %s, original %s", sc.ID, s.Status(), rec.Status)
		}
	}
}
