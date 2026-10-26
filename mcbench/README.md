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
go run ./cmd run -model google/gemma-4-26b-a4b-qat -instructions rules_agent,mcbench,strategy

# Scripted players - report them next to model results
go run ./cmd run -player random
go run ./cmd run -player heuristic

go run ./cmd report                 # every run under results/ and on the results branch; or pass run dirs / games.jsonl files
go run ./cmd analyze                # how the players behaved: loss causes, turn use, form, defence
```

Each run writes, per game, a JSON trace to `results/<run>/live/<game>.jsonl` (`<run>` is `<time>-<player>__<instructions>`)
(watch it with `go run ./cmd view -follow`), plus one record per finished game
in `games.jsonl` and the run's metadata (runner, machine, commit, settings) in
`run.json`. Progress is logged as structured `slog` to stderr. A finished run
with no `infra_error` game is then published to the `results` branch; `report`,
`analyze`, `view` and `board_gui` read local and published runs together by
default (`-source hybrid`; [details](../docs/architecture.md#records-and-live-status)).

To see a game as a board in the browser, step by step (arrow keys, Home/End or
the buttons), run the `board_gui` command from `mcbench/` and open
http://127.0.0.1:8090/:

```bash
go generate ./boardgui                                   # build the page (React, in boardgui/web/) once; needs Node and npm
go run ./cmd board_gui                                   # follow the run that is newest under results/ at start
go run ./cmd board_gui results/<run> -game seed3         # open a game of one run
go run ./cmd board_gui results/<run>/live/<game>.jsonl   # open one trace file
```

The command line only picks what opens first: the page's pickers open any run
and game under the results directory, and the address (`?run=…&game=…#step`)
keeps the choice for reload and links. A game still being played is followed as
its trace grows. Without a game picked, the page follows the run: it shows the
earliest-started game still being played and moves on to the next game when
that one ends (with `-parallel`, it stays on one game until it ends). Clicking a
card opens it full size; the side panel lists the current round's decisions
with the player's reasoning (and, for a model, its thinking). `board_gui` only
reads: it rebuilds each board by replaying the recorded choices on the same
scenario and seed, and stops with an error if the replay no longer matches the
trace (the engine or data changed since it was recorded). Card images load from
marvelcdb.com by card code; a card without one is drawn as text.

The runs on the `results` branch are also online at
https://kanareklife.github.io/ResearchProject/: the board, and a statistics
page that filters and groups every recorded game and aggregates it as `report`
and `analyze` do. Only a local run can be followed. The `Pages` workflow rebuilds it
on each push to `main` or `results`, with the `site` command, which writes the
same static site from any results directories:

```bash
go generate ./boardgui && go run ./cmd site -out site   # -source as for report; or list results directories
```

Configuration lives in `config.yaml` (grouped into `data`, `inference`,
`player`, `run`, `online`, `scoring`); flags override it. Key flags:

When `mcbench` starts, it automatically loads an optional `.env` file from the
current directory. Existing environment variables are kept, so shell exports
override values from `.env`. For example, from `mcbench/`:

```dotenv
OPENAI_BASE_URL=https://opencode.ai/zen/go/v1
OPENAI_API_KEY=your-key
```

The `.env` file is ignored by Git. Do not commit API keys.

- `-player model|systemone|heuristic|random|first` (`systemone` asks a System One classifier, e.g. `-model openjev-latest`, to pick an option from the board; it gets no instruction documents unless `-instructions` names some, e.g. `rules_short`), `-model`, `-server`, `-retries`, `-stream` (stream the replies of any `-api`, for gateways that cut slow requests),
  `-temperature`, `-max-tokens`, `-max-failures`, `-header "Name: value"` (repeatable; see below)
- `-instructions rules_agent,mcbench,strategy` and `-instructions-dir data/instructions`
- `-seeds N`, `-samples K`, `-parallel N`, `-out results`, `-publish=false` (keep the run local)
- `-pick 1,2,10`: play only these seeds, for example to repeat the games of a run that ended in `infra_error`
- `-source hybrid|local|online` (`report`, `analyze`, `view`, `board_gui`, `site`): which results to read (default `hybrid`)
- `-resume <run dir>` continues a model run that was cut short (quota, crash, kill): finished games are kept, each unfinished one is replayed from its trace to the same board and conversation and played on. The run's own settings apply unless a flag overrides them, so `-server` moves the rest of the run to another provider; the games keep the original model label. The instruction documents must be unchanged. `run.json` lists each resume.
- `-log-level debug|info|warn|error` (default `info` for `run`): structured `slog` output on stderr. `info` logs every game action and every choice, tagged with scenario, seed and sample; `debug` adds each model request; `warn` shows only rejected choices, API retries and games that end abnormally
- `OPENAI_BASE_URL` overrides the endpoint for every protocol, below `-server`; `inference.server_env` names a different variable. `OPENAI_API_KEY` is the key (`inference.api_key_env`)

`publish` merges the runs of one player and instruction set and publishes them
as one run, for example a run and the retries of its failed games. Per game it
keeps the latest result that is not an `infra_error`; failed games are dropped.
It refuses a run with fewer finished games than planned unless `-partial` is
given, and a run with no graded game at all. It fetches the results branch
before it pushes, never force-pushes, and retries a lost push race:

```bash
go run ./cmd publish results/<run> results/<retry-of-failed-seeds>
```

`-header` adds a request header to every API call. In its value, `{run}` (the
run's directory name, unique per run) and `{game}` (scenario, seed and sample,
e.g. `spider-man-vs-rhino_seed2_s0`) are filled per game, so each game can get
its own id. For a gateway that groups requests by session:

```bash
go run ./cmd run -model <id> -header "x-opencode-session: {run}_{game}"
```

Without placeholders the value is the same for the whole run. Any other
`{...}` in a value is rejected.

## Docs

- [How it works](../docs/architecture.md): packages, engine, the tool contract,
  the agentic loop, scoring, records
- [Creating scenarios](../docs/scenarios.md): YAML content, the card schema and
  effect vocabulary
- [Glossary](../docs/glossary.md): Marvel Champions and benchmark terms

## Current content

The Core Set starter decks of Spider-Man (Justice), Iron Man (Aggression) and
Captain Marvel (Leadership), each with its obligation and nemesis set, against
Rhino (with Bomb Scare), Klaw (with Masters of Evil) and Ultron (with Under
Attack), each villain at stages I–II with the Standard set: nine scenarios
(`go run ./cmd list`) over 10 seeds each. `run.scenarios` in `config.yaml`
selects the ones `run` and `play` use. Decks and scenarios are data under
`data/`; adding more needs no Go changes as long as the cards' rules exist.

Card behaviour follows the card text, the Rules Reference and its FAQ
(Appendix IV); effects on other players (Whirlwind's attack on each other hero)
do nothing solo.
