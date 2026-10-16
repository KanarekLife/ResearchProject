# Instruction documents

Each `<name>.md` here can be given to a model with `-instructions name[,name...]`; the documents are concatenated in the order given. This README is not an instruction document.

| Name | Role | Source |
|---|---|---|
| `rules` | rules base | Engine-aligned summary of the solo rules as mcbench implements them. |
| `rules_reference` | rules base | Close paraphrase of the [Rules Reference v1.8](../../../docs/references/rules_reference_v18.pdf), covering the whole solo game (not deck building, game modes, campaigns or multiplayer). Headings follow the Rules Reference glossary. |
| `learn_to_play` | rules base | Close paraphrase of the [Learn to Play](../../../docs/references/learn_to_play.pdf) book in its own order, solo only (no deck building, starter deck or scenario lists). |
| `rules_agent` | rules base | Rules written for an agent from both references, organized by the decisions a player makes, with the threat and damage arithmetic spelled out. Same scope as `rules_reference`. |
| `community` | rules base + tips | Tutorial-style explanation and recurring advice from community guides (see below). |
| `mcbench` | add-on | The mcbench tool and decision contract: tools, decision kinds, option keys, statuses, reply limits ([architecture](../../../docs/architecture.md)). No game rules. |
| `strategy` | add-on | General solo strategy, derived from the rules and the Learn to Play strategy tip. No card- or scenario-specific advice. |

Use one rules base and any add-ons, for example `rules_reference,mcbench,strategy`. `learn_to_play,rules_reference` gives both official books, the way they are meant to be read.

The `strategy` used by runs recorded before 2026-10-16 was a shorter document; records tell the two apart by the instruction hash.

## Sources of `community`
Paraphrased; tips tied to specific cards or villains are left out, and claims that contradict Rules Reference v1.8 were dropped or corrected.

- Zatu Games, "How to Play Marvel Champions" (M. Thomasson, 2019, updated 2022): https://zatu.com/blogs/how-to-play/how-to-play-marvel-champions
- Card Gamer, "Learning the Basics of Marvel Champions" (R. Sainio, 2025): https://cardgamer.com/games/living-card-games/marvel-champions/learning-the-basics-of-marvel-champions/
- Marvel Champions Monthly, "SHIELD Rookie Series: First Steps" (2020): https://marvelchampionsmonthly.podbean.com/e/shield-rookie-series-first-steps-your-first-game
- Alter-Ego Chronicle, "Tips and Strategies to Improve Your Game" (2023, updated 2025): https://alteregochronicle.wixsite.com/alter-ego-chronicle/post/tips-and-strategies-to-improve-your-game
- iRebelDestiny, "Marvel Champions Top 10, Issue #5: Tips For New(ish) Players" (2021): https://irebeldestiny.com/2021/03/05/marvel-champions-top-10-issue-5-tips-for-newish-players/
- The Ludologists, "10 Rules I Missed During First Playthroughs" (W. Lyles, 2021): https://www.ludologists.com/marvel-champions-10-rules-i-missed-during-first-playthroughs-and-you-may-be-too/
- Hall of Heroes, "Very Quick Marvel Champions Villain Tips" (2021, updated 2022): https://hallofheroeslcg.com/2021/02/22/very-quick-marvel-champions-villain-tips/
- FFG forum archive, "Please teach me how to win" (2020): https://ffg-forum-archive.entropicdreams.com/topic/308289-please-teach-me-how-to-win
- FFG forum archive, "Your approach to Solo Expert" (2019–2020): https://ffg-forum-archive.entropicdreams.com/topic/303366-your-approach-to-solo-expert
- MarvelCDB, "Beginner's guide: Core Set Only" (neothechosen, 2022): https://marvelcdb.com/decklist/view/21259/beginner-s-guide-core-set-only-1.0
