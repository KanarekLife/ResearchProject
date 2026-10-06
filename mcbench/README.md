# mcbench

A benchmark that measures how well a language model plays full solo games of
**Marvel Champions: The Card Game**, given only the instruction documents we
provide. The rules engine is a small Go reimplementation of the solo game
(Rules Reference v1.8), modelled on
[z00lus/marvel-lcg](https://github.com/z00lus/marvel-lcg), not a general
card-game platform.

Each game starts from regular setup with a fixed seed and is played to the end
by a player that talks to the game through a small set of deterministic tools.
A game is then scored from its final state on win, villain damage dealt, hero HP
left, final threat and speed.

## Quick start

```bash
GOTOOLCHAIN=auto go test ./...      # determinism, every seed playable, scoring, rules, contract
go run ./cmd list                   # scenarios
go run ./cmd validate               # play every seed with the scripted players

# A model through any OpenAI-compatible API (LM Studio on :1234 by default).
# The model plays by calling the game tools; see config.yaml for defaults.
go run ./cmd run -model google/gemma-4-26b-a4b-qat -instructions rules,strategy

# Scripted players - report them next to model results
go run ./cmd run -player random
go run ./cmd run -player heuristic

go run ./cmd report results/*/games.jsonl
```

Each run writes, per game, a human-readable board to
`results/<run>/live/<game>.txt` (tail it to watch) and a JSON trace to
`<game>.jsonl`, plus one record per finished game in `games.jsonl`. Progress is
logged as structured `slog` to stderr.

Configuration lives in `config.yaml` (grouped into `data`, `inference`,
`player`, `run`, `scoring`); flags override it. Key flags:

- `-player model|heuristic|random|first`, `-model`, `-server`, `-retries`,
  `-temperature`, `-max-tokens`, `-max-steps`
- `-instructions rules,strategy` and `-instructions-dir data/instructions`
- `-seeds N`, `-samples K`, `-parallel N`, `-out results`
- `-log-level debug|info|warn|error` (default `info` for `run`): structured `slog` output on stderr. `info` logs every game action and every choice, tagged with scenario, seed and sample; `debug` adds each model request; `warn` shows only rejected choices, API retries and games that end abnormally
- `OPENAI_BASE_URL` / `OPENAI_API_KEY`: override the endpoint and key

## Docs

- [How it works](../docs/architecture.md): packages, engine, the tool contract,
  the agentic loop, scoring, records
- [Creating scenarios](../docs/scenarios.md): YAML content, the card schema and
  effect vocabulary
- [Glossary](../docs/glossary.md): Marvel Champions and benchmark terms

## Current content

Spider-Man with his Justice starter deck (with Eviction Notice and the Vulture
nemesis set) against Rhino (stages I–II) plus the Standard and Bomb Scare
encounter sets, over 10 seeds. Decks and scenarios are data under `data/`;
adding more needs no Go changes.

Known simplifications:

- Spider-Sense draws automatically.
- Always-beneficial optional responses (Interrogation Room, Daredevil) trigger
  automatically.
- Great Responsibility applies only in the villain phase.
- Star boost abilities are ignored.
- Main schemes have a single stage.
