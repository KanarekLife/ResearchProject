# How mcbench works

## Layout

```
mcbench/game/
  engine/      rules only: the step machine, damage, threat, combat, payments
  cards/       card YAML and the effect interpreter
  scenario/    loads content and builds engine games
  session/     the player<->game contract: Session, View, live view
mcbench/player/
  player.go    Player interface and Usage
  heuristic/   Random, First, Heuristic (scripted players)
  human/       a player at the terminal
  instruction/ loads the instruction documents
  model/       the LLM player: agentic loop, the JSON tool contract and its
               dispatch, system prompt
mcbench/integrations/inference/
  inference.go            the Client interface and Message/Tool/Reply types
  openai_compatible/      the HTTP implementation
mcbench/benchmark/  the harness: play, score, record, report
mcbench/cmd/        the CLI and its config.yaml
mcbench/constants/  string literals shared across packages
mcbench/data/       scenarios/, decks/, villains/, encounter-sets/, cards/, instructions/
docs/               this file, scenario authoring, glossary
```

Dependencies point one way: `cmd -> benchmark -> player -> session ->
{scenario, cards, engine}`, `player/model -> integrations/inference`, and
`player/model -> game/cards` (the `get_card` lookup).
**The game never imports the player or the integrations.** The model player
reaches the game state only through `session.Session`, so the same game can be
played by a model, a script or a person. Only `cmd` reads `config.yaml` and
environment variables; every other package receives its settings as arguments.

## The engine

`game/engine` is a step machine. Rules are small steps pushed on a stack. `Run`
executes steps until one of them needs a choice, then stops with a **pending
decision** (a prompt plus legal options). `Choose(id)` runs that option's effect
and continues until the next decision or the end of the game:

```
Game.Run      pop steps until a decision is pending, the game is over, or halted
Game.Ask      pause on a Decision (options numbered here)
Game.Choose   run the chosen option, then Run again
Game.Do       push steps so steps[0] runs first
```

- Every legal option is listed up front, including card payments: one option per
  minimal way to pay. Illegal moves are impossible.
- All randomness comes from one generator seeded at setup. **The same seed and
  the same choices always give the same game.**
- `StartGame` shuffles, starts in alter-ego form, draws, offers a mulligan, then
  begins round 1.
- A game ends when the villain's last stage is defeated (win), the hero is
  defeated or the main scheme reaches its threshold (loss), or the round or
  decision limit is reached.

The package is split by concern: `game.go` is the step machine and `setup.go`
starts a game; `damage.go` handles statuses, damage and healing; `threat.go`
places and removes threat; `combat.go` handles attacks, thwarts and targeting;
`payment.go` enumerates payments; `player.go` runs the player phase and
`options.go` builds the options it offers; `villain.go` runs the villain phase,
`enemy_attack.go` resolves enemy attacks and `reveal.go` the encounter deck;
`naming.go` labels cards for option keys.

## The contract: game tools

A `session.Session` is one game in progress — the game side of the contract.
The player side is five deterministic JSON tools, the only way a model touches
the game. Both live in `player/model`: `tools.go` declares the schemas,
`call.go` executes them against the session. The game never sees a tool.

| Tool            | Arguments                         | Returns                                                                                                                                                   |
| --------------- | --------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `get_state`     | none                              | The public state (`session.View`): hero, hand, cards in play, pile sizes, villain and next stages, schemes, minions. Never deck order or face-down cards. |
| `get_decision`  | none                              | `{"status", "decision": {"kind", "prompt", "options": [{"id", "text", "key"}]}}`. The decision is absent once the game is over.                           |
| `choose_option` | `option_id`, optional `reasoning` | `{"status", "events": [...], "decision": {...}}`: what happened, then the next decision.                                                                  |
| `get_log`       | optional `last` (default 20)      | The last lines of the game log.                                                                                                                           |
| `get_card`      | `name`                            | A card's type, cost/stats and text.                                                                                                                       |

- **Status** (`session`, from `constants`): `awaiting_decision`, `won`, `lost`,
  `round_limit`, `decision_limit`, `engine_error`. The harness adds
  `agent_error`.
- **Decision kinds** (`constants.Kind*`): `mulligan`, `turn`, `defend`,
  `window` (an interrupt such as Backflip), `choice`, `discard`.
- **Option keys** are stable engine labels (`constants.Prefix*`), e.g.
  `play:Swinging Web Kick>Rhino|pay:Genius+Energy`. Option ids are 1..n in the
  order shown.
- Errors (such as an unknown option id) return `{"error": "..."}` and count as
  invalid choices.

Go players use the same contract as typed methods: `View()`, `Decision()`,
`Choose(id, reasoning)`, `Log(n)`. The tools are thin JSON wrappers.

**Option order** is shuffled per decision (seeded by game seed and sample,
never by player), so a player's position bias cannot leak into its score.

## The player and the agentic loop

A `player.Player` answers every decision of one game through the session
contract. There are three implementations: `heuristic` (scripted,
deterministic), `human`, and `model`.

`player/model.Model` plays a game as one conversation with an
`inference.Client`. There is one protocol: **the model calls the game tools
itself.** The fixed system prompt is embedded (`player/model/system.md`); the
instruction documents are appended at runtime and are the variable under study.

```
system: embedded format prompt + the instruction documents
for each decision:
    user: events since the last decision + the current decision (get_decision)
    model: tool calls (get_state / get_card / get_log / choose_option)
    the model player runs them against the session
until the status is no longer awaiting_decision
```

- **Context:** the conversation is append-only and sent in full every request:
  system prompt, every turn, every reply including its reasoning, and every tool
  result. Nothing is trimmed or rewritten, so each request's prompt starts with
  the previous request's prompt and the provider's prompt cache can reuse it.
- **Retries:** a decision may take up to `max_steps` model requests. Malformed
  answers and invalid options get an error reply and the model tries again. If
  it still fails, the game ends as `agent_error` with score 0.

## The inference client

`integrations/inference` defines the `Client` interface and the `Message`,
`Tool`, `ToolCall` and `Reply` types; it knows nothing about the game.
`integrations/inference/openai_compatible` implements it for any server that
speaks the OpenAI chat-completions shape (LM Studio, Ollama, llama.cpp, vLLM or
a hosted provider). It is a plain HTTP layer configured by a `Config` value the
caller builds; it reads no files and no environment variables. It:

- retries transient failures (429/5xx/network) with backoff, honoring
  `Retry-After`;
- sends `max_tokens`, or falls back to `max_completion_tokens` when a provider
  asks for it;
- reads reasoning from either `reasoning_content` or `reasoning`, and reports
  cached and reasoning tokens when the provider includes them.

`config.yaml` (loaded by `cmd`, flags override) holds the endpoint, sampling,
run size and scoring weights.

## Cards are data

Cards are YAML documents under `data/cards/` (embedded at build time), compiled
into engine `Script` hooks at load. A card lists its printed stats and a list of
abilities; each ability is a trigger plus a list of declarative **effects**
interpreted by `game/cards/effects.go`, with predicates and targeting in
`conditions.go` (verbs such as `attack`, `thwart`,
`draw`, `heal_villain`, `place_threat`, `choose`) gated by named **predicates**
(`hero`, `paid:mental`, `host_is_target`, ...). A handful of cards whose
behaviour cannot be expressed as a short effect list use a single-purpose helper
(`nemesis`, `assign_damage` for Explosion, `find_and_reveal`, `absorb_damage`).
See [scenarios](scenarios.md) for the schema.

## Scoring

When a game ends, `benchmark` scores its final state (`benchmark/score.go`).

| Criterion        | Meaning (each in [0, 1])                             | Default weight |
| ---------------- | ---------------------------------------------------- | -------------- |
| `win`            | 1 if the villain was defeated                        | 0.40           |
| `villain_damage` | share of the villain's total HP (all stages) removed | 0.25           |
| `hero_hp`        | share of the hero's HP left                          | 0.10           |
| `threat`         | 1 − main-scheme threat / threshold                   | 0.10           |
| `speed`          | for wins: 1 − (rounds − 1) / max rounds              | 0.15           |

The score is the weighted mean. The weights are the benchmark definition and
can be set in `config.yaml` under `scoring`; the defaults above are used when
none are given. **Never change the weights between runs you intend to
compare.** Reports show every criterion and the score (mean ± standard error),
plus rounds to win, decisions, invalid choices and tokens.

## Records and live status

`cmd run` logs structured progress with `log/slog` and writes to
`results/<time>-<player>/`:

- **`live/<scenario>_seed<N>_s<K>.txt`** — a human-readable board, appended
  after every decision: the action taken, its reasoning, the events and the
  resulting state. Tail it to watch a game.
- **`live/<scenario>_seed<N>_s<K>.jsonl`** — the same game as one JSON line per
  event (messages, choices, start, end).
- **`games.jsonl`** — one record per finished game: scenario, seed, player and
  instruction hash, status, stats, criteria, score, usage, every choice and the
  full log.

`cmd report` summarizes `games.jsonl` files.

## Logging

All logging is structured `log/slog` on stderr, installed once by `cmd`
(`-log-level` on `mcbench run`; every other command logs only warnings). A game
logs through its session's logger, which is already tagged with `scenario`,
`seed` and `sample`, so parallel games stay distinguishable.

| level   | what                                                                                                                              |
| ------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `info`  | every game action (the lines of the game log, with `round`) and every choice (`n`, `kind`, `key`), plus run and game start/finish |
| `debug` | each model request: decision, step, token counts, truncation                                                                      |
| `warn`  | a rejected choice, a failed API attempt, a game that ended abnormally                                                             |

The game log shown to the player is separate and unchanged: logging reads it, it
never feeds back into the game, so it cannot affect determinism or scores.

## Tests

`go test ./...` covers what a result depends on: determinism and every seed
finishing (`benchmark`), the scoring definition (`benchmark/score_test.go`),
rule interactions (`game/cards`), the answer contract (`game/session`: invalid
options, limits; `player/model`: tool errors and that every advertised tool
executes), the agentic
loop against a scripted client (`player/model`) and the HTTP client's retries
(`integrations`). Card YAML is
validated when it loads, so an unknown predicate fails at startup rather than
mid-game.
