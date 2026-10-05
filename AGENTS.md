# AGENTS.md

Guidance for coding agents working in this repository.

## Project goal

A benchmark (`mcbench/`) that measures how well an LLM plays **Marvel Champions: The Card Game** when given only the instruction documents we provide. Each game is a full solo game of a **scenario** (hero deck + villain) from a fixed seed, played to the end and scored from its final state.

The rules follow <https://github.com/z00lus/marvel-lcg> (Rules Reference v1.8, solo).

Read `mcbench/docs/` first:
- `architecture.md`: engine, contract, agentic loop, scoring
- `scenarios.md`: how to add content
- `glossary.md`: the terms to use in code, data, docs and conversation

## Priorities

- **Research value over implementation.** The engine is a tool. Judge every change by what it adds to the measurement: can a model play well from the provided instructions alone, and is the measurement trustworthy? If a change doesn't produce or improve a measurement, skip it or ask the user first.
- **Simple solutions.** No overcomplication, no speculative abstractions, no configuration nobody asked for. Prefer deleting code to adding it.
- **Small, sensible packages** with one job each. Keep this layering:
  - `engine/`: rules only; knows nothing about players
  - `cards/`, `scenario/`: content
  - `game/`: the contract
  - `agent/`: players
  - `bench/`: harness
  - `cmd/mcbench/`: CLI
- **Tests only where they can fail.** Test real behaviour that could regress: determinism, every seed playing to the end, rule interactions, answer parsing. Don't write tests for things that always pass.
- **Go, standard library only.** Go is pinned in `mise.toml`. Use the upstream Python repo as a reference for rules and card data; don't run or wrap it.
- **Port only what scenarios need.** Add a card or rule only when a scenario uses it, and check its behaviour against Rules Reference v1.8 and upstream.

## The agent <-> game contract

- Players (LLMs, scripted baselines, humans) use only `game.Session`: the deterministic tools `get_state`, `get_decision`, `choose_option`, `get_log` and `get_card`, or their typed Go equivalents. Players never import or touch `engine/`. The game must stay playable without AI.
- The contract is documented in `mcbench/docs/architecture.md`, and its schemas live in `game/tools.go` (`mcbench tools`). Update both together.

## Validity rules (do not regress)

- The public view never contains deck order, encounter deck contents or face-down cards.
- Option order is shuffled per decision, seeded by game seed and sample, never by player. Report the `random` and `heuristic` baselines next to model results.
- The same seed and choices must give the same game (`bench.TestDeterministic`). Every scenario seed must play to the end (`bench.TestGamesComplete`, `mcbench validate`).
- Scoring weights (`bench/score.go`) are part of the benchmark definition. Never change them between runs that will be compared.
- Instruction documents are the independent variable. The harness prompt in `agent/llm.go` explains only the answer format, never game knowledge.
- The default model endpoint is local LM Studio (`http://localhost:1234/v1`) through the OpenAI-compatible API. Never mix models within one comparison.

## Git workflow

- Remote: `git@github.com:KanarekLife/ResearchProject.git`.
- Use [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`), with an optional scope such as `feat(engine): ...`.
- Never commit or push to `main` directly. Work on a branch and open a pull request with `gh pr create`. Don't merge it yourself unless the user asks.
- Never commit `results/`, `bin/` or logs.

## Notes on upstream

- Upstream ships no LICENSE file. Don't copy its code or data verbatim, and check licensing before publishing anything derived from it.
- Upstream's `docs/engine_architecture.md`, `docs/card_scripting_guide.md` and `unit_test/test_v18_*` tests are useful references for exact timing and rules behaviour.
