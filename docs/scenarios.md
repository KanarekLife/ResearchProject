# Creating a scenario

All content lives under `mcbench/data/`. A scenario is a matchup: a **hero
deck** against a **villain** with its encounter sets. Cards are referenced by
name, or by code when a name is ambiguous (Rhino's stages all share one name).

```
data/scenarios/<id>.yaml
data/decks/<id>.yaml
data/villains/<id>.yaml
data/encounter-sets/<id>.yaml
data/cards/<set>.yaml      # the card definitions, not listed by hand
```

Adding decks and scenarios needs no Go changes; only new cards need a card
definition.

## 1. Hero deck: `data/decks/<id>.yaml`

```yaml
id: spider-man-justice
name: Spider-Man (Justice starter deck)
identity: Spider-Man
cards:
  - Black Cat
  - Backflip
  # ... 40-50 cards: the hero's signature cards plus aspect and basic cards
obligation: Eviction Notice
nemesis:
  [
    Vulture,
    Highway Robbery,
    Sweeping Swoop,
    Sweeping Swoop,
    The Vulture's Plans,
  ]
```

The obligation is shuffled into the encounter deck at setup. The nemesis set is
set aside until _Shadow of the Past_ brings it in.

## 2. Villain: `data/villains/<id>.yaml`

```yaml
id: rhino
name: Rhino
stages: ["01094", "01095"] # I, II, ... (standard uses I and II, expert II and III)
main_scheme: The Break-In!
cards:
  - Armored Rhino Suit
  - Charge
  # ...
```

## 3. Encounter sets: `data/encounter-sets/<id>.yaml`

```yaml
id: bomb-scare
name: Bomb Scare
cards:
  [Bomb Scare, Hydra Bomber, Hydra Bomber, Explosion, False Alarm, False Alarm]
```

The standard set and modular sets share this format.

## 4. Scenario: `data/scenarios/<id>.yaml`

```yaml
id: spider-man-vs-rhino
title: Spider-Man (Justice) vs Rhino - Standard, Bomb Scare
deck: spider-man-justice
villain: rhino
encounter_sets: [standard, bomb-scare]
seeds: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
max_rounds: 20
max_decisions: 1000
```

- `seeds` fix the shuffles. Every player plays the same seeds, so results are
  comparable.
- `max_rounds` ends unfinished games (scored as not won); `max_decisions`
  guards against engine bugs.

## Card definitions: `data/cards/<set>.yaml`

Every card is a YAML document in a list under `cards:`. Stats are copied from
the printed card; `text` is our own paraphrase (players see it, so it must
describe what the code does). Loading fails with `unknown card "..."` until a
referenced card is defined.

```yaml
cards:
  - code: "01005"
    name: Swinging Web Kick
    type:
      event # hero, alter-ego, ally, event, support, upgrade, resource,
      # villain, main_scheme, side_scheme, minion, treachery,
      # attachment, obligation
    aspect: Hero # Hero, Justice, Aggression, Leadership, Protection, Basic
    cost: 3
    resources: [mental] # physical, mental, energy, wild
    text: "Hero Action (attack): deal 8 damage to an enemy."
    abilities:
      - trigger: play # play, action, reveal, defeated, interrupt, forced, resource, stat, boost
        when: [hero] # predicates; the card can only be played while they hold
        target: enemy # selector for the effect's target
        effects:
          - { verb: attack, damage: 8 }
```

An ability is a trigger plus declarative effects. `when` gates the whole
ability; every effect may have its own `when`. A `boost` ability is a card's
"Boost:" text: it runs for each friendly character damaged by the activation
the card boosts (`event_target` is that character). Effects support branches:
`if_zero`, `if_already`, `if_empty`, and `choose` with named `options`.

**Effect verbs** (the values of `verb`): `attack`, `damage_enemy`, `damage_hero`,
`thwart`, `remove_threat`, `place_threat`, `draw`, `heal`, `heal_hero`, `heal_villain`,
`stun`, `stun_chosen`, `confuse`, `give_tough`, `exhaust`, `ready`,
`flip_alter_ego`, `discard_random`, `discard_random_place_threat`, `surge`,
`scheme`, `villain_attack`, `villain_and_minions_attack`, `remove_from_game`,
`reveal`, `shuffle_encounter`, `detach`, `to_encounter_discard`, `cancel`,
`prevent_damage`, `reduce_amount`, `take_threat_as_damage`, `absorb_damage`,
`counter`, `cost_reduction`, `mill_keep`, `find_and_reveal`, `take_random_card`,
`return_attached`, `assign_damage`, `nemesis`, `choose`, `discard_chosen`.

**Predicates** (`when`): `hero`, `alter_ego`, `hero_damaged`, `hero_confused`,
`villain_tough`, `schemes`, `minions`, `not_exhausted`, `exhausted`,
`counters_positive`, `hero_exhausted`, `hero_ready`, `source_self`,
`host_is_source`, `host_is_target`, `bomb_scare`, `not_bomb_scare`, `vulture`,
`upgrades_supports`, `amount_positive`, `paid:<resource>`,
`not_paid:<resource>` (a wild resource counts as any type).

**Selectors** (`target`): `enemy`, `all_enemy`, `minion`, `scheme`,
`upgrade_support`, `enemy_without_webbed_up`, `damaged_character`,
`event_source`, `event_target`, `chosen`, `hero`, `villain`, `self`.

The vocabulary lives in `mcbench/constants`; the interpreter is
`game/cards/effects.go`. Prefer an existing verb; a new one is a small case
there. A few cards (`Shadow of the Past`, `Explosion`, `Rhino II`, `Armored
Rhino Suit`, `Highway Robbery`, `The Vulture's Plans`) use a single-purpose
helper in the same file.

## Checking a scenario

```bash
go run ./cmd list                                # does it load?
go run ./cmd validate -only <id>                 # every seed, played by the scripted players
go run ./cmd play -only <id> -seed 3             # play it yourself
go run ./cmd run -player heuristic -only <id>    # scripted numbers to compare models against
```

`validate` must report no engine errors. Read a few games with `go run ./cmd view` to
confirm the new cards behave as their text says.
