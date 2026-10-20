# Instruction documents

Pass documents to a model with `-instructions name[,name...]`; they are appended to the system prompt in that order. This README is not one of them. Use one rules base and any add-ons, for example `rules_agent,mcbench,strategy`. Records identify the exact text by the instruction hash.

| Name | Role | Source |
|---|---|---|
| `rules_reference` | rules base | Close paraphrase of the [Rules Reference v1.8](../../../docs/references/rules_reference_v18.pdf), solo game only. Headings follow its glossary. |
| `learn_to_play` | rules base | Close paraphrase of the [Learn to Play](../../../docs/references/learn_to_play.pdf) book in its own order, solo only. |
| `rules_agent` | rules base | Rules written for an agent from both references, organized by the decisions a player makes, with the threat and damage arithmetic spelled out. |
| `rules_short` | rules base | Seven bullets: the win and loss conditions, the two forms, paying for cards, the villain phase and defending. Written for the `systemone` player, which has no room for the long documents. |
| `mcbench` | add-on | The tool and decision contract: tools, decision kinds, statuses, reply limits ([architecture](../../../docs/architecture.md)). No game rules. |
| `strategy` | add-on | General solo strategy from the rules, the Learn to Play strategy tip and the play of winning games ([research](../../../docs/research.md)). No card- or scenario-specific advice. |
| `algorithm` | add-on | A step-by-step if-then procedure for every decision kind, built from the decisions of won games ([research](../../../docs/research.md)). Limits are formulas from the board, not fixed numbers. No card- or scenario-specific advice. |

`rules_reference` and `learn_to_play` follow the official books and are not edited for strategy; `learn_to_play,rules_reference` gives both books the way they are meant to be read.
