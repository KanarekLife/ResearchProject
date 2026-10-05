# How mcbench works

## Packages

```
engine/    the rules: setup, turns, payments, combat, villain phase; knows nothing about players
cards/     card definitions (stats from the printed cards, text paraphrased, scripted effects)
scenario/  loads scenarios (deck + villain + encounter sets) and builds engine games
game/      the contract: a Session, its tools and the public View
agent/     players: LLM (agentic loop), scripted baselines, human
bench/     the harness: plays games, scores them, writes records and reports
cmd/mcbench  the CLI
```

Each package depends only on the ones above it. Players (`agent/`) talk to the game only through `game.Session`, never to the engine directly. That's why the same game can be played by a model, a script or a person.

## The game engine

The engine is a step machine. Rules are written as small steps on a stack. `Run` executes steps until one of them needs a choice, then stops with a **pending decision** (a prompt plus legal options). `Choose(option)` runs that option's effect and continues until the next decision or the end of the game.

- Every legal option is listed up front, including card payments: one option per minimal way to pay. Illegal moves are impossible.
- All randomness (shuffles, random discards) comes from one generator seeded at setup. **The same seed and the same choices always give the same game.**
- `StartGame` performs setup: shuffle the decks (your obligation goes into the encounter deck), start in alter-ego form, draw your hand, offer a mulligan, then begin round 1.
- A game ends when the villain's last stage is defeated (win), the hero is defeated or the main scheme reaches its threshold (loss), or the scenario's round limit is reached.

## The contract: game tools

A `game.Session` is one game in progress. Its contract is five deterministic tools: the same game history always gives the same results. Their JSON schemas come from the code (`mcbench tools`).

| Tool | Arguments | Returns |
|---|---|---|
| `get_state` | none | The public state (`game.View`): hero, hand, cards in play, deck size, discard pile, villain and next stages, main and side schemes, minions, encounter deck size and discard pile. Never deck order or face-down cards. |
| `get_decision` | none | `{"status", "decision": {"kind", "prompt", "options": [{"id", "text", "key"}]}}`. The decision is absent once the game is over. |
| `choose_option` | `option_id`, optional `reasoning` | `{"status", "events": [...], "decision": {...}}`: what happened, then the next decision. |
| `get_log` | optional `last` (default 20) | The last lines of the game log. |
| `get_card` | `name` | A card's type, cost/stats and text. |

- **Status** is one of `awaiting_decision`, `won`, `lost`, `round_limit`, `decision_limit` or `engine_error`.
- **Decision kinds:**
  - `mulligan`: setup
  - `turn`: your actions on your turn
  - `defend`: an enemy attacks you
  - `window`: you may play an interrupt such as Backflip
  - `choice`: a card effect asks you to choose
  - `discard`: end of turn
- **Errors** (such as an unknown option id) are returned as `{"error": "..."}` and counted as invalid choices.

Go players use the same contract as typed methods: `View()`, `Decision()`, `Choose(id, reasoning)`, `Log(n)`. The tools are thin JSON wrappers around these methods.

**Option order** is shuffled per decision (seeded by game seed and sample, never by player). The engine lists options in a fixed order (cards first, end turn last), so without shuffling a player's position bias would leak into its score.

## The agentic loop

`agent.LLM` plays a game as one conversation with an OpenAI-compatible model. A local LM Studio server is the default.

```
system: answer format (+ instruction documents, the variable under study)
for each decision:
    user: events since the last decision + the current decision (get_decision)
          [json protocol: + the state (get_state)]
    model: ... → choose_option(option_id, reasoning)
    the session runs the game to the next decision
until the status is no longer awaiting_decision
```

The model can answer in two ways (`-protocol`):

- **tools**: the model receives the tool schemas and calls tools itself. It may call `get_state`, `get_card` and `get_log` as often as it likes before `choose_option`. Every call goes to `Session.Call`.
- **json**: for models without function calling. The harness calls `get_state` and `get_decision` for the model and puts the results in the prompt. The model answers `{"reasoning": ..., "option_id": N}` and the harness passes it to `choose_option`.

Both protocols reach the game only through the tools.

- **Context:** the conversation is append-only and sent in full on every request: system prompt, every turn, every reply including its reasoning (`reasoning_content`), and every tool result. Nothing is trimmed or rewritten, so each request's prompt starts with the previous request's prompt and the server's prompt cache can reuse it. Very long games can therefore reach the model's context limit, which ends the game as `agent_error`. The chat template decides what the model actually sees: LM Studio's Gemma template drops earlier reasoning when rendering, so for Gemma it is stored but not read.
- **Retries:** a decision may take up to `-max-steps` model requests. Malformed answers and invalid options get an error reply, and the model tries again. If it still fails, the game ends as `agent_error` with score 0.

## Scoring

When a game ends, `bench` scores its final state. The calculation is deterministic and lives in `bench/score.go`.

| Criterion | Meaning (each in [0, 1]) | Weight |
|---|---|---|
| `win` | 1 if the villain was defeated | 0.40 |
| `villain_damage` | share of the villain's total HP (all stages) removed | 0.25 |
| `hero_hp` | share of the hero's HP left | 0.10 |
| `threat` | 1 − main-scheme threat / threshold | 0.10 |
| `speed` | for wins: 1 − (rounds − 1) / max rounds | 0.15 |

The score is the weighted mean. Reports show every criterion and the score (mean ± standard error over games), plus rounds to win, decisions per game, invalid choices and tokens. Never change the weights between runs you intend to compare.

## Records

`mcbench run` writes to `results/<time>-<agent>/`:

- **`live/<scenario>_seed<N>_s<K>.jsonl`**: written as the game happens, one JSON line per event, so a running or crashed game can be inspected (`tail -f`).
  - `start`: scenario, seed, agent, instruction set.
  - `message`: one conversation message (system, user, assistant with `reasoning_content`, tool result). Assistant messages also carry `prompt_tokens`, `completion_tokens` and `truncated`.
  - `choice`: the decision made (round, kind, option key and text, reasoning), the events that followed, and the new status.
  - `end`: status, error, rounds, criteria, score, token usage.
- **`games.jsonl`**: one record per finished game. It holds the scenario, seed, agent and instruction set (names and content hash), the status, final stats, criteria and score, token usage, every choice, and the full game log.

`mcbench report` summarizes `games.jsonl` files.
