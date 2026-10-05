# Glossary

Marvel Champions terms, as used in this project's code, data and reports.

## Game structure
- **Scenario**: a matchup. A hero deck plus a villain with its main scheme and encounter set, plus the standard and modular encounter sets shuffled in. Example: `spider-man-vs-rhino`.
- **Hero deck**: the identity card, the player deck (40 cards: the hero's 15 signature cards plus aspect and basic cards), the hero's obligation, and the nemesis set.
- **Identity**: the double-sided hero card. **Hero form** (Spider-Man) has ATK/THW/DEF; **alter-ego form** (Peter Parker) has REC. You may change form once per turn.
- **Aspect**: the class of a player card. Aggression, Justice, Leadership, Protection, Basic, or Hero (the hero's signature cards).
- **Villain**: the main enemy, played in **stages** (I, II, III). Defeating a stage reveals the next one with full HP; defeating the last stage wins the game.
- **Main scheme**: the villain's plan. Threat accumulates on it; if it reaches its **threshold**, you lose.
- **Encounter deck**: the villain's deck (villain set + standard set + modular set + your obligation). Each villain phase deals you encounter cards.
- **Encounter set**: a group of encounter cards. The **villain set** belongs to the villain; the **standard set** is in every scenario; a **modular set** (e.g. Bomb Scare) adds variety.
- **Nemesis set**: your hero's personal enemies (Spider-Man's is Vulture). It is set aside at setup and enters play through *Shadow of the Past*.
- **Obligation**: a hero-specific encounter card (Spider-Man's is *Eviction Notice*) that forces a choice when revealed.

## Rounds and phases
- **Round**: one player phase followed by one villain phase.
- **Player phase / turn**: you take any actions, then end your turn. You draw up to your **hand size**, then ready your exhausted cards.
- **Villain phase**: threat is placed on the main scheme; the villain then **activates** (it attacks you in hero form, or schemes against your alter-ego), and so does each engaged minion; finally encounter cards are dealt and revealed.
- **Mulligan**: at setup, discard any cards from your opening hand and draw back up.

## Cards and resources
- **Ally**: a character that attacks, thwarts or defends for you, taking **consequential damage** after it attacks or thwarts. Limit 3.
- **Event**: a one-shot effect that is then discarded. **Support** and **upgrade** cards stay in play.
- **Resource**: physical, mental, energy or wild. You pay a card's cost by discarding other cards, each providing its printed resources.
- **Exhaust / ready**: turning a card sideways to use it; it is readied at the end of your player phase.

## Enemies and threat
- **Minion**: a smaller enemy engaged with you that attacks or schemes alongside the villain.
- **Treachery**: a one-time encounter effect. **Attachment**: an encounter card attached to the villain.
- **Side scheme**: an extra scheme. It is defeated when its threat reaches 0. Icons: **acceleration** (+1 threat each round), **hazard** (+1 encounter card each round), **crisis** (main-scheme threat cannot be removed).
- **Threat**: tokens on schemes. **Thwart**: remove threat. **Scheme**: place threat.
- **Boost**: when the villain activates, a face-down encounter card adds its boost icons to the attack or scheme.
- **Surge**: reveal one more encounter card.

## Keywords and statuses
- **Guard**: while a guard minion is engaged with you, you cannot attack the villain.
- **Toughness / tough**: a tough status prevents the next damage.
- **Stunned**: the next attack is replaced by discarding the stun. **Confused**: the next thwart or scheme is replaced by discarding the confuse.
- **Overkill**: excess damage to an ally carries over to the hero. **Quickstrike**: a minion attacks as soon as it engages you.

## Benchmark terms
- **Seed**: fixes all shuffles of a game; scenarios list their seeds, and every agent plays the same ones.
- **Sample**: one of several games on the same seed (useful at temperature > 0).
- **Decision**: one point where the player must choose. **Option**: one legal choice, numbered in the order shown. **Option key**: a stable engine label for an option (e.g. `play:Swinging Web Kick>Rhino|pay:Genius+Energy`), used by scripted players and in records.
- **Session**: one game in progress, played through the contract (see [architecture](architecture.md)).
- **Contract / tools**: the deterministic tools a player uses: `get_state`, `get_decision`, `choose_option`, `get_log`, `get_card`.
- **Player / agent**: anything that plays a session: a language model, a scripted baseline or a human.
- **Agentic loop**: a game played as one model conversation, one turn per decision (the last `-history` turns are kept). `-history 0` makes every decision independent.
- **Instruction set**: the documents given to the model (`instructions/*.md`); the main variable under study.
- **Criteria / score**: the per-game measurements and their weighted mean (see [architecture](architecture.md#scoring)).
- **Round limit**: games still running after `max_rounds` stop and are scored as not won.
