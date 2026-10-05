# Creating a scenario

A scenario is a matchup: a **hero deck** against a **villain** with its encounter sets. It's made of four kinds of JSON files, and cards are referenced by name. When a name is ambiguous (Rhino's stages all share one name), use the card code from `cards/`.

## 1. Hero deck: `decks/<id>.json`

```json
{
  "id": "spider-man-justice",
  "name": "Spider-Man (Justice starter deck)",
  "identity": "Spider-Man",
  "cards": ["Black Cat", "Backflip", "Backflip", "..."],
  "obligation": "Eviction Notice",
  "nemesis": ["Vulture", "Highway Robbery", "Sweeping Swoop", "Sweeping Swoop", "The Vulture's Plans"]
}
```

`cards` is the 40–50 card player deck: the hero's signature cards plus aspect and basic cards. The obligation is shuffled into the encounter deck at setup. The nemesis set is set aside until *Shadow of the Past* brings it in.

## 2. Villain: `villains/<id>.json`

```json
{
  "id": "rhino",
  "name": "Rhino",
  "stages": ["01094", "01095"],
  "main_scheme": "The Break-In!",
  "cards": ["Armored Rhino Suit", "Charge", "Charge", "..."]
}
```

`stages` lists the villain stages in order; standard mode uses I and II, expert mode II and III. `cards` is the villain's own encounter set.

## 3. Encounter sets: `encounter-sets/<id>.json`

```json
{"id": "bomb-scare", "name": "Bomb Scare", "cards": ["Bomb Scare", "Hydra Bomber", "Hydra Bomber", "Explosion", "False Alarm", "False Alarm"]}
```

The standard set (`standard`) and modular sets use this format.

## 4. Scenario: `scenarios/<id>.json`

```json
{
  "id": "spider-man-vs-rhino",
  "title": "Spider-Man (Justice) vs Rhino - Standard, Bomb Scare",
  "deck": "spider-man-justice",
  "villain": "rhino",
  "encounter_sets": ["standard", "bomb-scare"],
  "seeds": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
  "max_rounds": 20,
  "max_decisions": 1000
}
```

- `seeds` fix the shuffles. Every agent plays the same seeds, so results are comparable. More seeds give tighter error bars at the cost of more games.
- `max_rounds` stops games that run too long; they score as not won.
- `max_decisions` guards against runaway games (an engine bug).

## Implementing missing cards

Loading fails with `unknown card "..."` when a card isn't implemented. Add only the cards your scenario needs:

1. Take the stats and text from the upstream data (`data/cards.json` in [z00lus/marvel-lcg](https://github.com/z00lus/marvel-lcg)), and check the behaviour against the Rules Reference v1.8.
2. Add a `CardDef` in `cards/` with the printed stats and a short paraphrase in `Text`. Players see that text, so it must describe what the code actually does.
3. Script the ability with the engine's existing hooks (`OnPlay`, `OnReveal`, `Forced`, `Actions`, `PlayWindow`, `ResourceAbility`, ...); the cards already in `cards/` show each one.

## Checking a scenario

```bash
go run ./cmd/mcbench list                                # does it load?
go run ./cmd/mcbench show -only <id>                     # the opening state and decision
go run ./cmd/mcbench validate -only <id>                 # every seed, played by the baselines
go run ./cmd/mcbench play -only <id> -seed 3             # play it yourself
go run ./cmd/mcbench run -agent heuristic -only <id>     # baseline numbers to compare models against
```

`validate` must report no engine errors. Read a few game logs in the records to confirm the new cards behave as their text says.
