# Research notes

How language models play Marvel Champions in mcbench, what separates the games
they win from the games they lose, and what we changed in response. This is a
living document: add to it when a run teaches something new, and keep each claim
tied to something a reader can recheck with `mcbench report` (scores) and
`mcbench analyze` (behaviour) on the run directories under `mcbench/results/`.
Game terms are in the [glossary](glossary.md).

Everything comes from one scenario (Spider-Man vs Rhino), mostly ten seeds or
fewer, one game per seed. Read it as well-founded hypotheses, not settled results.

## In short

- Games are won by **tempo**: build the board early, attack every turn, use every
  card and ally, and keep the hero in hero form. Games are lost by **passivity**.
- The strongest models (Gemini 3.8 Flash, gpt-6.1-sol, and as a first impression
  Gemini 4 Argon) play exactly that way. The Claude models (Haiku, Sonnet) play the
  opposite: they hold cards, defend with the hero and sit in alter-ego.
- That passivity did not go away with longer rules documents or more reasoning
  effort. A strategy worded as a mandatory order, or a step-by-step play algorithm
  built from won games, moves what Haiku does partway and wins it a few games, but
  it still plays far from the winners.
- Simple scripted players (random, first option, a heuristic) lose every game, so
  winning needs real play, not luck.

## How games are won

**Tempo.** Won games are fast and busy. The winning player develops its board in
the first rounds (allies, supports, upgrades), attacks from its first turn as a
hero, and takes the villain's first stage down early. Every round the villain
survives adds another villain phase, more threat and another encounter card, so a
slow game becomes a lost one even when no single turn looks bad.

**Using everything, every turn.** Winners end their turns with nothing left over:
the hero has used its basic power, every ally has attacked or thwarted, cards in
play have used their actions and few playable cards stay in hand. Losers end turns
with ready allies and full hands.

**The hero attacks; the rest of the board thwarts.** Winners spend the hero's
basic power on attacks almost every turn, even with some threat on the scheme, and
remove threat with allies, supports and events. The hero thwarts only when the
scheme is genuinely close to completing.

**Hero form as the normal state.** Winners spend nearly all their turns in hero
form. In alter-ego the villain schemes instead of attacking, with a boost card on
top, and that typically adds several times the threat of a hero-form round.
Winners change to alter-ego rarely, and almost never with the scheme anywhere near
its target.

**Taking small hits.** Winners let small attacks through, or block with an ally,
so that their hero stays ready. A hero that defends stays exhausted through the
next turn and loses its basic power.

What did not predict wins: aggressive mulligans (they helped one model, another
won without them), and visible planning in a model's reasoning (the model that
talked most about threat won nothing).

## How each model plays

Behaviour is read with `mcbench analyze`: how many things a model does before it
ends its turn, how often it blocks with the hero (which then skips its next basic
power), and how often the hero attacks or thwarts. Samples are small, so treat
models with a handful of games as a first impression.

**Gemini 3.8 Flash** plays the winning pattern. It attacks every turn, uses its
whole board, almost never blocks with the hero, and ends turns with few playable
cards left in hand. Its weakness is the alter-ego change: nearly all of its losses
came from ending a turn in alter-ego, where the villain's scheme plus a high boost
completed the main scheme. It sometimes knows the risk and takes it anyway, and once
stunned the villain just before changing form, which wasted the stun.

**gpt-6.1-sol and Gemini 4 Argon** play the same way: a lot of actions per turn and
almost nothing left unused. gpt-6.1-sol is the most convincing of the newer models,
and its loss was a long game. Argon has barely been tried, so it is a first
impression only.

**DeepSeek Flash** is as aggressive as Gemini and understands that defending costs
the next turn. It overdoes it: it almost never retreats to alter-ego, even at low
hit points against the stronger second villain stage, so most of its losses are its
hero being defeated. It also leaves more cards unplayed than Gemini. The later
DeepSeek V4.1 Flash shows the same weakness. Playing aggressively is necessary but
not enough: the player also has to know when to back off.

**Mistral Large 4** has the right habits (it attacks and uses its board) but takes
fewer actions per turn than the winners, and it ends many turns in alter-ego. Two
games lost to the scheme.

**Claude models (Haiku, Sonnet, Opus)** play defensively. They block with the hero
against most attacks, judging only the damage saved and not the lost attack, so the
hero rarely gets to act. Haiku also lingers in alter-ego early, where the villain's
scheming can end the game within a few rounds, and its turns are short, ending with
ready allies and playable cards. Its own justifications read like loss aversion:
"Black Cat would cost Web Kick and Webbed Up; not worth it. Hold cards and end
turn." Haiku loses mostly because the scheme completes, often while its hero is
healthy. Sonnet behaves like a milder Haiku and at higher effort stays out of
alter-ego much more. Opus uses its turns best of the three (more actions, fewer
cards left in hand) and thwarts carefully, but it still blocks with the hero often
and in one game never damaged the second villain stage.

## Why Claude plays passively, and what moved it

What we tried on the Claude models:

- **Longer or different rules documents** (`rules`, `rules_reference`,
  `rules_agent`, with or without `strategy`): no change for Haiku.
- **More reasoning effort:** Haiku at high effort did the same things per turn as at
  medium, only with twice the thinking. Sonnet at high effort did better, mainly by
  staying out of alter-ego, but is still far from the winners.
- **A mandatory strategy text:** the `strategy` document given the framing of a
  standing order that overrides the model's own judgement, with a turn checklist
  and three hard rules. It shifted Haiku's play partway: more actions per turn, more
  basic powers, fewer blocks with the hero, fewer turns ended in alter-ego. An early
  small test looked promising (Haiku won some games), but a larger repeat won
  none, and most losses were still to the scheme. The habits changed only part
  of the way: about three actions per turn against the winners' six, half of the
  attacks still blocked with the hero, and more than half of turns ending with a
  playable card in hand. The early wins were most likely luck. Taking the text
  apart (the framing alone, the checklist alone, the hard rules alone) did not
  isolate a part that carries the effect either. One repeat game shows the failure
  well: Haiku held the scheme off for eighteen rounds and never once damaged the
  villain, so the villain simply outlasted it.
- **Making the model plan a whole turn in one go** (a single tool call for the full
  turn, tried on Gemini 3.8 Flash) made play worse. The model cannot see the result
  of one action before committing the next. When offered only as an option, it was
  almost never used.
- **A play algorithm** (`algorithm`): an if-then procedure for every decision kind,
  taken from what won games did (set up in alter-ego and attack as a hero in round
  one, attack while the scheme is safe, take small hits instead of defending with the
  hero, end in alter-ego only at very low threat). On Haiku at high effort it won
  about as often as `strategy` in the same run, where Haiku with neither won
  nothing. Haiku follows the rules of the procedure that are easy to check:
  it attacks in round one in every game and ends far fewer turns in alter-ego than
  with `strategy`. It does not follow the ones that need judgement: about a third of
  its turns still end with a card it could play, and its wins take ten rounds or
  more. Variants did not do better: a short version and a version with fixed
  thresholds instead of formulas won less, and a version that made Haiku write out
  the threat and damage numbers before ending a turn or defending did not win at all.
- **The full official rule books** (a near-verbatim text of Learn to Play and the
  Rules Reference, more than ten times the length of `rules_reference`; tried once
  and not kept) made Haiku more passive than any other set: it blocked with the hero
  in two of three attacks, used fewer basic powers, did not win and cost about five
  times the tokens of `rules_agent` with `algorithm`. More rules text is not what
  Haiku lacks.

So Claude's passivity looks like a disposition more than a knowledge gap or a lack
of thinking time. Wording that tells the model not to hold back moves its behaviour;
wording that only informs it does not. Even with the behaviour moved, Haiku still
loses, so it is probably missing something else: the judgement about when an attack,
a thwart or a card is worth more than keeping the hero safe.

## Instruction documents

- Differences between instruction sets were within the noise, and the
  sets we compared changed several documents at once, so no effect could be
  attributed to one document.
- The `mcbench` document, which explains the tools, removed nearly all of the
  protocol mistakes models made without it.
- A strategy text can do harm. An earlier `strategy` presented defending with the
  hero as "keeping it ready", the reverse of the rules, and the most defensive model
  followed it closely. Strategy advice must be checked against the rules as
  carefully as the rules documents themselves.
- A model that already plays well is not hurt by the mandatory text: Gemini 3.8
  Flash keeps winning most games with `strategy`, `algorithm` or both, and the
  differences between the three are within the noise.
- Short texts do worse than long ones: a short `algorithm` and a short `strategy`
  won less than the full versions on Gemini, and the short `algorithm` never won on
  Haiku. The detail seems to matter more than the framing.
- Retrieval over the instructions (giving the model only the relevant rules) is
  unlikely to help as things stand. The prompts are short and cached, models
  rarely state a wrong rule (about one hero defence in a hundred cites a wrong
  readying rule), and the rules Haiku breaks are rules it was given. The one clear
  case of a missing rule (DeepSeek believing a defending hero readies anyway, in a
  quarter of its hero defences with `rules_reference`) mostly went away where
  `rules_agent` states the rule plainly.

Planned grid: every set includes `mcbench`; the rules base is none,
`rules_reference` or `rules_agent`; `strategy` is on or off. Six conditions, each
answering one question, with several games per seed.

## What the harness taught us

- Models look things up a lot, and the lookups must not count against them. Only
  failed replies (no tool call, or a tool error) now do.
- Models type card names loosely: escaped characters, partial names, alter-ego
  names, names shared by several villain stages. The card lookup now tolerates all
  of these.
- Some models send several choices in one reply, as if option numbers stayed valid
  after a choice. The tool description now says to choose once per reply.
- Listing every way to pay for every card made decisions long and hard to read. A
  card is now one option, and the payment is a separate decision that can be
  cancelled.
- Provider failures (rate limits, billing, server errors) are recorded as
  `infra_error` and not scored, so they no longer count as model failures.
- The temperature was never actually sent, so all runs used each provider's
  default; it is now explicit (`-1` means the provider default).
- Models showed at most a weak preference for the first option, so options are no
  longer shuffled: they come grouped by kind of action, which is easier to read and
  stays the same between similar decisions.
## Changes made in response

- `strategy` now leads with fast, aggressive play (build and attack early, attack
  by default, react only to real danger), plans for a high boost, spells out the
  cost of ending a turn in alter-ego, asks to use every ally and action, and states
  the real cost of defending.
- `rules_agent` states that a hero that defends cannot use a basic power on the
  next turn.
- `rules` and `community` were removed: one duplicated `rules_agent`, the other
  mixed rules with tips and blurred the effect of `strategy`.
- Each recorded choice now stores the hero's form and hit points, the threat and
  the options offered, which is what `mcbench analyze` reads.
- `algorithm` was added as an add-on next to `strategy`: the same lessons written
  as an ordered procedure, so the two framings can be compared.

## A classifier baseline

A System One classifier that sees the board, the legal options and the last few
log lines plays far below the instruction-driven models and near the scripted
heuristic. Neither a longer log nor a short rules summary changes that
consistently. It has no memory, so given a step that can be undone it repeats
it forever (play a card, back out of the payment, play it again) until the
decision limit; offering each payment as part of the card play removed the
loops without making it win.

## Open questions

- A mandatory strategy or a play algorithm only partly fixes Haiku's habits and
  results. Is it still missing the judgement of when to attack, thwart or hold? Do
  the texts do better on Sonnet?
- Would a reminder of the matching rule inside each decision (the defend rules on a
  defend decision) move Haiku further than the same rule in the system prompt?
  This changes the player <-> game contract and needs a decision first.
- Is the Claude models' defending a knowledge gap that documents can close, or a
  disposition they keep regardless? (The evidence so far points to disposition.)
- Do gpt-6.1-sol and Gemini 4 Argon keep winning across all ten seeds?
- Does planning for a high boost stop aggressive models like DeepSeek from dying at
  low hit points, without making them passive?
- Does the separate payment decision change error rates or token use?
- Some seeds are hard for every player. Should scores be read relative to the
  scripted baseline on the same seed?
- Would a second hero or scenario show the same ranking, or is the tempo pattern
  specific to Spider-Man vs Rhino?
