# Agent Instructions

`mcbench/` is a benchmark for one research question: how well can a language
model play a full solo game of Marvel Champions: The Card Game when the only
thing it is given is a set of instruction documents? The rules engine is a small
Go reimplementation of the solo game, not a general card-game platform. The
human researcher owns the methodology; agents implement the agreed scope.

See `mcbench/AGENTS.md` for Go conventions.

## Approval gate

Benchmark and research decisions belong to the human. Get explicit approval
before changing any of these, because they define what the benchmark measures
and make past runs incomparable:

- the scoring criteria or their weights (`benchmark/score.go`);
- the instruction documents or the harness system prompt;
- scenario, deck, villain or encounter-set content, and scenario seeds;
- the player <-> game contract (tools, statuses, decision kinds, option keys);
- how seeds or options are ordered (determinism).

Present the proposed change, its scope, the rationale and the alternatives, then
wait for approval. Do not infer approval from a ticket, an earlier discussion,
an implementation request or silence. If a request hides an unstated benchmark
decision, stop and ask before making it.

## Guardrails

- Work only on the requested scope. No unrelated cleanup, refactor or
  future-proofing.
- Prefer the smallest clear solution; delete code rather than add it.
- Never invent rules, card text, sources or results. Card behavior follows Rules
  Reference v1.8 and the upstream implementation; a card's text is our own
  paraphrase of what the code does.
- Keep benchmark and training data traceable to its source. Never copy upstream
  code or data verbatim, and check licensing before publishing derived content.
- Preserve determinism: the same seed and the same choices must give the same
  game. `TestDeterministic` and `TestGamesComplete` are the guard for this and
  must stay green.
- Keep work concise and readable. If the human cannot see why something exists,
  it is not ready.
- All rules should be following the official Marvel Champions rules available [here](./docs/references/rules_reference_v18.pdf).
- Source of truth for cards in available in the [marvelcdb.com](https://marvelcdb.com/).

## Commits

- Conventional Commits (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`,
  `chore:`) with an imperative summary; one logical change per commit.
- Do not commit unless the human asks. Never commit `results/`, `bin/` or logs.
- No `Co-Authored-By` trailers or other AI attribution.
- Never commit to `main`; work on a branch.

## Documentation

- Update a document only when an approved decision changes what it describes.
- Keep documentation human-readable and research-focused: no filler, no
  restating code, no generic best-practices text.
- Keep each fact in its owning document and link instead of copying it.
  `docs/architecture.md` owns how the engine, contract, agentic loop and scoring
  work.

## Verify before done

From `mcbench/`:

```bash
GOTOOLCHAIN=auto go test ./...   # determinism, every seed finishes, rule checks
go run ./cmd validate            # play every seed with the scripted players
```

`validate` is the smoke test: any `engine_error` or `decision_limit` it reports
is a bug.
