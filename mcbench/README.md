# mcbench

A benchmark that measures how well a language model plays full solo games of **Marvel Champions: The Card Game**, given only the instruction documents we provide. The rules engine is a minimal Go reimplementation (Rules Reference v1.8, solo only), based on [z00lus/marvel-lcg](https://github.com/z00lus/marvel-lcg).

Each game starts from regular setup with a fixed seed and is played to the end by an agent that talks to the game through a small set of deterministic tools. Games are scored from the final state on win, villain damage dealt, hero HP left, final threat and speed.

## Quick start

```bash
go test ./...                       # determinism, every seed playable, rule checks
go run ./cmd/mcbench list           # scenarios
go run ./cmd/mcbench tools          # the agent <-> game contract
go run ./cmd/mcbench validate       # play every seed with the baselines

# A model through any OpenAI-compatible API (LM Studio on :1234 by default)
go run ./cmd/mcbench run -model google/gemma-4-26b-a4b-qat -instructions rules,strategy

# Baselines - report them next to model results
go run ./cmd/mcbench run -agent random
go run ./cmd/mcbench run -agent heuristic

go run ./cmd/mcbench report results/*/games.jsonl
```

Useful `run` flags:
- `-protocol tools`: the model calls the game tools itself (default `json`)
- `-history N`: decisions kept in the conversation; `0` makes every decision independent
- `-seeds N`: play only the first N seeds
- `-samples K`: games per seed, with `-temperature > 0`
- `-parallel N`: games run concurrently
- `OPENAI_BASE_URL` / `OPENAI_API_KEY`: use another endpoint

## Docs

- [How it works](docs/architecture.md): engine, game tools contract, agentic loop, scoring, records
- [Creating scenarios](docs/scenarios.md): decks, villains, encounter sets, adding cards
- [Glossary](docs/glossary.md): Marvel Champions and benchmark terms

## Current content

Spider-Man with his Justice starter deck (with Eviction Notice and the Vulture nemesis set) against Rhino (stages I–II) plus the Standard and Bomb Scare encounter sets, over 10 seeds.

Known simplifications:
- Spider-Sense draws automatically.
- Always-beneficial optional responses (Interrogation Room, Daredevil) trigger automatically.
- Great Responsibility applies only in the villain phase.
- Star boost abilities are ignored.
- Main schemes have a single stage.
