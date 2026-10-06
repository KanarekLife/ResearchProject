# mcbench — Agent Instructions

Go conventions for the benchmark code. The research guardrails in the repository
root [AGENTS.md](../AGENTS.md) — approval gate, scope, commits and documentation
policy — also apply here.

## Architecture constraints

- Dependencies point one way:
  `cmd -> benchmark -> player -> game/session -> {scenario, cards, engine}`, and
  `player/model -> {integrations/inference, game/cards}`. The game never imports
  the player or the integrations.
- The JSON tool contract (schemas and `Call`) belongs to `player/model`; the
  game exposes only the typed methods.
- A player reaches the game state only through `session.Session`, so the same
  game can be played by a model, a script or a person.
- Only `cmd` reads `config.yaml` or environment variables. Every other package
  receives its settings as arguments.
- All randomness comes from the engine's seeded RNG. Do not call `math/rand`
  outside the engine and the players' own choice RNG.

## Code style

- Every package has a one-sentence `// Package <name> ...` doc comment on its
  primary file.
- No magic strings. Put a literal in `mcbench/constants` when more than one
  package needs it; otherwise declare it as a named constant in the file that
  uses it. Literal JSON/YAML struct tags are the one exception.
- Prefer small packages and small files with one job. Split a function when it
  starts doing more than one thing, and prefer a clear name to a clever one.
- Write a comment only where it gives value (a "why", an invariant, a gotcha).
  Do not restate what the code already says.

## Dependencies

- Go standard library plus `goccy/go-yaml`. Get explicit approval before adding
  anything to `go.mod`.

## Testing

- Standard `testing`, no framework.
- Add a test only where it can catch a regression: determinism, every seed
  finishing, rule interactions, answer parsing. Do not assert a constant against
  itself, a trivial constructor's output, or embedded prompt text.

## After implementing

- Once a change works and is tested, review it for refactor and simplification
  before calling it done.
- `gofmt`, `go vet ./...` and `GOTOOLCHAIN=auto go test ./...` must pass.
