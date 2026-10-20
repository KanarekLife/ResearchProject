# How mcbench works

## Layout

```
mcbench/game/
  engine/      rules only: the step machine, damage, threat, combat, payments
  cards/       card YAML and the effect interpreter
  scenario/    loads content and builds engine games
  session/     the player<->game contract: Session, View
mcbench/player/
  player.go    Player interface and Usage
  heuristic/   Random, First, Heuristic (scripted players)
  human/       a player at the terminal
  instruction/ loads the instruction documents
  model/       the LLM player: agentic loop, the JSON tool contract and its
               dispatch, system prompt
  systemone/   the System One classifier player
mcbench/integrations/systemone/   System One client (a classifier, not a chat model)
mcbench/integrations/inference/
  inference.go            the Client interface and Message/Tool/Reply types
  openai_compatible/      the HTTP implementation
mcbench/benchmark/  the harness: play, score, record, report, analyze
mcbench/cmd/        the CLI and its config.yaml
mcbench/constants/  string literals shared across packages
mcbench/ui/         terminal panels and colour, used by `human` and the trace viewer
mcbench/boardgui/   a trace as a board in the browser (`cmd board_gui`)
mcbench/data/       scenarios/, decks/, villains/, encounter-sets/, cards/, instructions/
docs/               this file, scenario authoring, glossary, research notes
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

- Every legal option is listed, including every minimal way to pay (in a
  `payment` decision when there are several). Illegal moves are impossible.
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
| `get_state`     | none                              | The public state, compacted from `session.View`: hero, hand, cards in play, pile sizes (not the discard lists), villain, schemes, minions; no later villain stages (see `get_card`); empty, zero and false fields left out. Never deck order or face-down cards. |
| `get_decision`  | none                              | `{"status", "decision": {"kind", "prompt", "options": ["1) text", ...]}}`. The decision is absent once the game is over.                                |
| `choose_option` | `option_id`, optional `reasoning` | `{"status", "events": [...], "decision": {...}}`: what happened, then the next decision.                                                                  |
| `get_log`       | optional `last` (default 20)      | The last lines of the game log.                                                                                                                           |
| `get_card`      | `name`                            | A card's type, cost/stats and text. Matches codes and names, ignoring case, punctuation and HTML escapes, then partial names and small typos; several matches return their codes. |

- **Status** (`session`, from `constants`): `awaiting_decision`, `won`, `lost`,
  `round_limit`, `decision_limit`, `engine_error`. The harness adds
  `agent_error` (the model failed) and `infra_error` (the model could not be
  reached).
- **Decision kinds** (`constants.Kind*`): `mulligan`, `turn`, `defend`,
  `window` (an interrupt such as Backflip), `choice`, `discard`, `payment`.
- **Option keys** are stable engine labels (`constants.Prefix*`), e.g.
  `play:Swinging Web Kick>Rhino`. They are for traces and scripted players and
  are not sent to the model: it sees each option as `N) text` (ids are 1..n in
  the order shown), because the decision text stays in the conversation and is
  resent on every later request.
- **Forced decisions:** a decision with a single legal option (the only target
  of an effect, or "End your turn" when nothing else is possible, about 8% of
  decisions) is applied by `session.Choose` itself, so the player is never asked
  it. Each is still recorded in `Choices` (`auto: true`), traced, logged at info
  level (`auto=true`) and shown in the events as `Auto-resolved (only legal
  option): ...`. `Options.KeepForced` turns this off; the board GUI replay sets
  it so that it applies every recorded choice. The model's `choose_option`
  result also leaves out the `Player chose: ...` line, which only repeats its
  own choice.
- **Payments:** an action with one way to pay (or none to make) is one option
  whose key carries it (`play:Aunt May|pay:Energy`). An action with several is
  one option followed by a `payment` decision: one `pay:Genius+Energy` option
  per minimal payment and `cancel`, which returns to the previous decision with
  nothing spent. A payment's text is its sources and their resources, e.g.
  `Genius (mental,mental), Peter Parker ability (mental)`, since there can be
  dozens of them. This keeps the turn list to one entry per card and target.
- Errors (such as an unknown option id) return `{"error": "..."}` and count as
  invalid choices.

Go players use the same contract as typed methods: `View()`, `Decision()`,
`Choose(id, reasoning)`, `Log(n)`. The tools are thin JSON wrappers.

**Option order** is the engine's, grouped by kind of action: on a turn, card
plays, hero powers, ally actions, card actions, form change, then end turn; in
a payment decision, `cancel` comes last. Models showed at most a weak
preference for early positions, so a readable, stable order was chosen over a
shuffled one ([research](research.md)).

## The player and the agentic loop

A `player.Player` answers every decision of one game through the session
contract. There are four implementations: `heuristic` (scripted,
deterministic), `human`, `model` and `systemone`.

`player/model.Model` plays a game as one conversation with an
`inference.Client`. There is one protocol: **the model calls the game tools
itself.** The fixed system prompt is embedded (`player/model/system.md`); the
instruction documents are appended at runtime and are the variable under study.

```
system: embedded format prompt + the instruction documents
user: the first decision (get_decision)
for each decision:
    model: tool calls (get_state / get_card / get_log / choose_option)
    the model player runs them against the session
until the status is no longer awaiting_decision
```

A successful `choose_option` result carries the events that followed and the
next decision, so it is the only place the model receives later decisions;
no user turn repeats it.

- **Context:** the conversation is append-only and sent in full every request:
  system prompt, every turn, every reply including its reasoning, and every tool
  result. Nothing is trimmed or rewritten, so each request's prompt starts with
  the previous request's prompt and the provider's prompt cache can reuse it.
- **Retries:** lookups (`get_state`, `get_card`, `get_log`, `get_decision`)
  that succeed are free. A reply fails when it has no tool call or a call
  returns an error (an invalid option, an unknown card); it gets an error reply
  and the model tries again. After `max_failures` failed replies for one
  decision, or 100 replies of any kind (a guard against endless loops), the game
  ends as `agent_error` with score 0.
- **Infrastructure failures:** when a request still fails after the client's
  retries (network, any HTTP error such as 402, 429 or 5xx, an unreadable
  response, or a cancelled run), the game ends as `infra_error`. It says
  nothing about the model, so reports count it but leave it out of every mean;
  rerun that seed.

**Resuming a game.** The trace of a game holds every message and choice, so a
cut-short game can be continued: `run -resume <dir>` keeps the finished games,
and the model player replays the recorded choices of each unfinished game
(restoring the board, since the same choices give the same game) and reloads its
conversation up to the last complete exchange. Only the endpoint may change, so a
run can move to another provider and keep one model label; the instruction
documents must be unchanged. `run.json` lists each resume.

**The System One player.** `systemone` is a baseline, not a model run in the
agentic loop. A System One endpoint is a classifier: given a state and a
multiple-choice question it returns a probability per option. For every decision
the player sends the compact board, the last log lines, the decision prompt and
the options, and plays the most probable one. It has no memory between
decisions, takes instruction documents only when asked (`rules_short` is written
for it), and sees a card with several payments as one option per payment, with
the cards discarded in its text, because a separate payment step with `cancel`
let it loop forever. The session's `Preview` finds those payments by replaying
the game on a copy.

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
- sends `reasoning_effort` (`-reasoning-effort`, a named level for hosted
  models) and `reasoning_budget_tokens` (`-reasoning-budget`, a thinking-token
  cap for llama.cpp-style servers) only when set. The setting the chosen API
  is actually sent is appended to the player name (`model:<id>@effort-high`,
  `@budget-4096`), so runs at different settings are reported separately;
  Gemini and Messages drop the budget when a level is set, and Responses
  takes no budget;
- reads reasoning from either `reasoning_content` or `reasoning`, and reports
  cached and reasoning tokens when the provider includes them;
- with `-stream` asks for the reply as a stream and assembles it into the same
  reply. A gateway that cuts a request after about two minutes (HTTP 524) kills
  a slow reasoning model on a long conversation, and every retry takes as long;
  a stream keeps the connection alive. Scoring and traces are unchanged.

`config.yaml` (loaded by `cmd`, flags override) holds the endpoint, sampling,
run size and scoring weights.

## Cards are data

All game content (cards, scenarios, decks, villains, encounter sets) is YAML
under `data/`, embedded at build time (`data.FS`), so a build carries exactly
the game it scores; only the instruction documents are read from disk. Cards are
YAML documents under `data/cards/`, compiled
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

- **`live/<scenario>_seed<N>_s<K>.jsonl`** — the game as one JSON line per
  event (messages, choices, start, end), appended as it happens. `cmd view`
  renders it; `cmd view -follow` watches a game in progress. The trace holds
  choices, not boards, so `cmd board_gui` rebuilds each board by replaying the choices
  on the same scenario and seed. A replay needs an engine that offers the same
  option keys: traces recorded before the `payment` decision do not replay.
- **`games.jsonl`** — one record per finished game: scenario, seed, player and
  instruction hash, status, stats, criteria, score, usage, every choice and the
  full log. Each choice also records the hero's form and HP, the main-scheme
  threat and the option keys offered at that moment.
- **`run.json`** — who ran it (git `user.name`), when, on what machine
  (hostname, OS, kernel, CPU, cores, memory, GPUs, Go version), from which
  commit and whether the tree was dirty, the run settings and scenarios, and how
  many games ended in each status.

A run started in the same second as one of the same name gets a `-2`, `-3`, ...
suffix. When it ends, `cmd run` publishes the run directory (unless
`-publish=false`) to the `results` branch of `origin`: one commit adds it as a
top-level directory, with the same suffixing if another run already holds the
name, and a push that loses a race is retried on the new branch tip. A run cut
short or with any `infra_error` game is not published, because reports leave
infra errors out and the run would be incomplete. A published local run records
its name on the branch in a `published` file. The branch is kept in a detached
worktree (`online-results/`) that only this step writes to.

`report`, `analyze`, `view` and `board_gui` read `-source local` (the results
directory), `online` (the branch, fetched first; offline, the last fetched copy)
or `hybrid` (both, the default). In `hybrid`, a published local run is read
once, and a different online run whose name is taken is listed as
`<run>@online-results`.

`cmd report` summarizes `games.jsonl` files, or every one under the directories
it is given (all of `-source` by default): an overview per scenario, then a
panel per player and instruction set with a row per game. Records of the same
player, instructions and scenario are pooled, even across runs.

`cmd analyze` reads the same files and reports behaviour instead of scores:
what the lost games were lost to, hero basic powers per round and how many
attacked, how often the hero defended, and how turns ended (in alter-ego, in
alter-ego near the target, with a basic power, an ally or a playable card left
unused). The turn-end measures need the per-choice board facts, so older
records show `-`. [Research](research.md) explains what the measures are for.

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
