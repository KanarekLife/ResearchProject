# Marvel Champions: The Card Game — rules reference (solo)

A condensed paraphrase of the official Rules Reference v1.8, limited to what can happen in a solo game (one player, one hero). Headings follow the Rules Reference glossary entries.

## Golden rules
- When a card's text or a scenario's rules contradict the rules, the card or scenario wins.
- "Cannot" is absolute: when two abilities or rules conflict, the one with "cannot" wins.
- Grim rule: if a rules or timing question has no answer, resolve it in the way that is worst for you, then continue.
- Tokens, counters and status cards are unlimited.

## Winning and losing
- You win when the final stage of the villain deck is defeated (its hit points reach 0).
- You lose when your identity is defeated (your hit points reach 0), or when the final stage of the main scheme is completed (its threat reaches its target threat).
- Some scenarios add other win or loss conditions; these cannot be cancelled.

## Card types
- **Player cards:** identity (hero on one side, alter-ego on the other), ally, event, support, upgrade, resource, player side scheme.
- **Encounter cards:** villain, main scheme, side scheme, minion, treachery, attachment, environment, obligation.
- **Characters** are identities, allies, villains and minions. **Enemies** are villains and minions.
- **Friendly** means cards you control.

## Setup (solo)
1. Put your identity into play alter-ego side up and set your hit points to the identity's printed hit points.
2. Set aside your obligation card and your nemesis set (out of play).
3. Shuffle your player deck.
4. Put the villain deck (stages in order, stage I on top) and the main scheme deck into play. Set the villain's hit points to the value on the current stage.
5. Shuffle the scenario's encounter sets together with your obligation to make the encounter deck.
6. Put any cards with the **setup** keyword into play. Resolve the main scheme's Setup abilities, flip it to side 1B (place its starting threat) and resolve its When Revealed abilities; then resolve the villain's Setup and When Revealed abilities.
7. Draw cards up to your hand size.
8. Mulligan: you may discard any number of cards from your hand, then draw back up to your hand size. (The discarded cards go to your discard pile; they are not shuffled back.)
9. Resolve any Setup abilities on your cards in play.

Player card abilities cannot be used during setup unless they have a "Setup" trigger.

## Round sequence
Each round has a player phase and then a villain phase.

### Player phase: your turn
In any order, and as often as you can pay for them:
- Change form (hero to alter-ego or back), **once per turn**.
- Play an ally, support, upgrade or player side scheme from your hand.
- Use your hero's basic attack or basic thwart (hero form) or your alter-ego's basic recovery (alter-ego form).
- Use a ready ally to attack an enemy or thwart a scheme.
- Trigger an Action ability on a card you control in play, on an encounter card in play, or on an event in your hand (by playing it). "Hero Action" and "Alter-Ego Action" need the matching form.

Every "Forced Action" must be resolved before the player phase can end (unless its cost cannot be paid or it has no valid target).

### End of the player phase
1. You may discard any number of cards from your hand, and must discard down to your hand size if you are above it.
2. Draw up to your hand size, one card at a time.
3. Ready all your exhausted cards (and any exhausted encounter cards).
4. Effects lasting "until the end of the phase" end; then "when the player phase ends" effects resolve.

### Villain phase
1. **Place threat.** Put the threat shown in the main scheme's acceleration field on the main scheme, plus 1 for each acceleration icon in play and each acceleration token.
2. **Enemies activate.** The villain activates once against you, then each minion engaged with you activates, one at a time in the order you choose. If you are in hero form, each enemy attacks you; in alter-ego form, each enemy schemes.
3. **Deal encounter cards.** You are dealt 1 face-down encounter card, plus 1 for each hazard icon in play.
4. **Reveal encounter cards.** Reveal and resolve the face-down encounter cards dealt to you, one at a time, in the order they were dealt. Cards dealt to you earlier in the round, or while these are resolving, join this queue.
5. The round ends: "until the end of the round" effects end, then "when the round ends" effects resolve.

## Identity and form
- You are always in exactly one form; the face-up side of the identity card is in play, the other side is out of play.
- You start the game in alter-ego form.
- You may change form once during each of your turns. A card effect that changes your form does not use up this change.
- Changing form changes only the form: damage, status cards, attached cards, tokens, lasting effects and ready/exhausted state all stay.
- Abilities that refer to your hero do nothing while you are an alter-ego, and the reverse.
- A card marked "[form] form only" can only be played or put into play in that form.
- Hero form gives basic attack (ATK), basic thwart (THW) and basic defense (DEF). Alter-ego form gives basic recovery (REC). Your hand size and some abilities differ between forms.
- Enemies attack a hero and scheme against an alter-ego.

## Resources and costs
- **Generating resources:** discard a card from your hand to generate the resource icons printed in its bottom-left corner, or use a "Resource" ability on a card you control.
- **Types:** energy, mental, physical and wild. A wild resource can be used as any type. Outside paying a cost, a wild resource is only "wild".
- Most costs accept any mix of types. Some cards and abilities require specific types, or have effects that check which types were paid.
- You may generate more than a cost requires; the excess is lost and does not count as paid.
- A cost with a per-player icon is multiplied by the number of players (1 in solo).
- If several costs are paid together (for example a card's cost and an additional cost), you pay them all at once or not at all.
- **Requirement (resources):** the card can only be played if the listed resources are among those spent.
- A card "put into play" ignores its cost and is not considered played. A card played "ignoring its resource cost" counts as played with 0 resources paid.
- If dealing damage is a cost, it is paid even if the damage is prevented. If taking damage is a cost, it is paid only if all the damage is taken.

## Playing cards and initiating abilities
1. Declare the card or ability. A played card is placed on the table (it is not yet in play).
2. Check play restrictions: timing, form requirements, and that every required target has at least one valid choice.
3. Determine the cost, including modifiers and the value of any X.
4. Pay the cost. If it cannot be paid, nothing happens and nothing is paid.
5. The card is played or the ability resolves. Allies, supports and upgrades enter play ready; an event resolves and then goes to your discard pile.

- An ability with several sentences is read in full first (for alteration effects such as "additional" or "each time"), then resolved one sentence at a time.
- Events, resources and upgrades on you are extensions of your identity: attacks, thwarts and defenses they make are made by your identity. Allies, supports, player side schemes and encounter cards are not.

## Ability types and timing
- **Constant abilities** (no bold trigger) are active while the card is in play.
- **Triggered abilities** have a bold trigger followed by a colon.
  - Optional: Action, Interrupt, Response, Resource.
  - Mandatory: Forced Action, Forced Interrupt, Forced Response, When Revealed, When Defeated, When Completed, Boost, Setup, and keywords. A "may" inside a mandatory ability makes only that part optional.
- **Action:** used during your turn.
- **Interrupt:** used when its trigger is about to happen, before it resolves; it can prevent or change the outcome. An interrupt with "would" happens earlier than one without.
- **Response:** used right after its trigger has happened.
- Each interrupt or response can be used once per occurrence of its trigger (each copy separately).
- **Resource:** used while you are paying a cost.
- **Special:** used only when another ability tells you to.
- **Labeled abilities** such as "Hero Action (attack)" or "(thwart)" or "(defense)" count as an attack, thwart or defense by your identity. Unless the text says so, using one does not exhaust your hero.
- **Timing priority** for abilities with the same trigger: constant abilities, delayed and lasting effects; then status cards; then Forced Interrupts; then Interrupts; then Boost and When Revealed abilities; then Forced Responses; then Responses.
- **Cancel:** the ability still counts as used and its costs stay paid; only its effects are prevented. A cancelled event still goes to the discard pile; a cancelled treachery still counts as revealed.
- **Replacement ("instead"):** the original event no longer happens and nothing can respond to it.
- **"Then":** the part after "then" happens only if the part before it fully resolved.
- **"Otherwise":** happens only if the preceding effect's condition was false or the effect could not resolve at all.
- **"And":** the joined effects happen together, independently; resolve as much of each as possible.
- **"For each":** without "choose" it is one instance applied to one target; with "choose" each iteration is a separate instance and may pick a different target.
- **Lasting effects** ("until the end of the phase") keep working even if their source leaves play, and update as the game state changes.
- **Delayed effects** happen automatically at their stated future time.
- **Limit X per [period]** applies to each copy; **Max X per [period]** applies across all copies.

## Targets
- An ability or basic power that needs a target can be started only if at least one valid target exists.
- A target is valid if any part of the ability can affect it.
- When a player card offers a choice of options, you cannot pick an option you cannot at least partly resolve.
- An ability that affects "each" of something can be used if at least one is a valid target.
- A character with 0 ATK, THW or SCH can still use that power against a valid target.

## Basic powers
- **Basic attack (ATK):** exhaust your hero to deal damage equal to ATK to one enemy that can be attacked.
- **Basic thwart (THW):** exhaust your hero to remove threat equal to THW from one scheme with threat on it.
- **Basic defense (DEF):** when you are attacked, exhaust your hero to defend; the attack's damage is reduced by DEF and the rest is dealt to your hero.
- **Basic recovery (REC):** exhaust your alter-ego to heal damage equal to REC. You cannot recover with no damage.
- A power whose value is a dash (–) cannot be used. Healing cannot take you above your maximum hit points.

## Allies
- Allies enter play ready. You may control at most **3 allies** (the ally limit). You may play another ally; you then immediately discard allies down to the limit, before the new ally's abilities resolve.
- During your turn, exhaust a ready ally to attack an enemy (dealing its ATK) or thwart a scheme (removing its THW).
- **Consequential damage:** after an ally attacks or thwarts, it takes damage equal to the number of consequential damage icons under its ATK or THW. It is taken after abilities triggered by that attack or thwart. If the target leaves play first, or the ally was stunned/confused, it takes none.
- An ally can defend you: exhaust it, and it takes all the damage from the attack. Excess damage does not carry over to you unless the attack has overkill.
- An ally whose damage reaches its hit points is defeated and discarded.

## Enemy activations
Each time the villain attacks or schemes, it gets one face-down **boost card** from the top of the encounter deck. Minions get no boost card unless they have the **villainous** keyword.

### Enemy attack
1. Give the villain (or a villainous minion) a face-down boost card.
2. Declare a defender: exhaust your hero (basic defense) or one of your allies, or do not defend.
3. Flip each boost card in the order dealt: resolve its "Boost" ability (marked by a star in the boost area; nothing else on the card is used), add its boost icons to the attacker's ATK, then discard it.
4. Calculate damage: the attacker's ATK with all modifiers and boost icons, reduced by the defending hero's DEF.
5. Deal the damage to the defending hero, to the defending ally (all of it), or, if undefended, to the attacked character (normally your hero).
6. Afterwards, "after [enemy] attacks" abilities, retaliate and similar forced abilities resolve, then optional ones.

- An attack is "initiated" as soon as the game determines it will happen; "when [enemy] attacks" interrupts resolve before step 1.
- If a defending ally leaves play before damage is dealt, the attack becomes undefended and hits your identity.
- If an activating minion leaves play, its activation ends at once.
- If the villain stage is defeated during its own activation, see Villain defeat.

### Enemy scheme
1. Give the villain (or a villainous minion) a face-down boost card.
2. Flip each boost card: resolve its Boost ability, add its boost icons to SCH, discard it.
3. Place threat on the main scheme equal to the enemy's modified SCH.

### Boost
- The star icon is not a boost icon and adds nothing by itself.
- Boost icons from several boost cards add up; an **amplify** icon in play adds one boost icon to each boost card.
- Damage from a Boost ability is not damage from the attack.
- A boost card given outside the enemy's activation stays face down on it until it next activates (a villain still gets its normal boost card then).

## Damage, healing and defeat
- Damage on an identity or villain lowers its hit point dial; damage on an ally or minion is tracked with damage tokens.
- A character with zero or fewer remaining hit points is defeated. Allies and minions go to their discard piles; a defeated identity or villain stage is removed from the game.
- **Prevented** damage reduces the damage taken, not the damage dealt. If all damage is prevented, the character has not taken damage.
- **Excess damage** is damage beyond a character's remaining hit points.
- **Overkill:** if an attack with overkill defeats an ally, the excess is dealt to that ally's controller's identity; if it defeats a minion, the excess is dealt to the villain.
- **Indirect damage** is divided as you choose among characters you control, no more on each than would defeat it.
- Moving damage off a character heals it; moving damage onto a character deals damage to it.
- "Gets +X hit points" raises hit points while active; if it ends and a character is left with no hit points, it is defeated.

## Threat and schemes
- Threat is placed on schemes and removed by thwarting and card abilities. When an ability places or removes threat on "a scheme", you choose the scheme.
- **Main scheme:** threat is checked continuously. When it reaches the target threat, that stage is completed and the main scheme deck advances: remove the stage, resolve the next card's side-A When Revealed, flip it to side B with its starting threat, resolve that side's When Revealed. Excess threat does not carry over; acceleration tokens do. Completing the final stage loses the game. Main schemes cannot be discarded.
- **Side schemes** enter play with their starting threat already on them (plus any **hinder** value). A side scheme with no threat is defeated and discarded.
- **Player side schemes** are side schemes you play; you may have only one in play in solo.
- **Crisis icon:** while one is in play, player cards cannot remove threat from the main scheme (encounter cards still can).
- **Acceleration icon:** +1 threat on the main scheme during each villain phase's place-threat step.
- **Hazard icon:** +1 encounter card dealt during each villain phase.
- **Amplify icon:** +1 boost icon on each boost card.
- **Acceleration token:** +1 threat on the main scheme each villain phase. Tokens on the main scheme cannot be removed.
- **Preventing threat** reduces the threat before it is placed.

## Encounter cards
- **Dealing** a card puts it face down in front of you; it is revealed in step 4 of the villain phase.
- When an encounter card is revealed:
  - **Minion:** enters play engaged with you.
  - **Treachery:** resolve it, then discard it.
  - **Attachment:** enters play attached to what its "attach to" text says (normally the villain).
  - **Side scheme:** enters play next to the main scheme.
  - **Environment:** enters play in the villain's play area and changes the rules while it stays.
  - **Obligation:** enters your play area; you resolve it as its text says.
  - Then resolve its When Revealed abilities (including incite and surge). Responses wait until the whole reveal is done.
- Abilities marked "Hero" or "Alter-Ego" on encounter cards resolve only if you are in that form.
- **Nemesis:** when a card brings in your nemesis, put your nemesis side scheme into play and your nemesis minion into play engaged with you; the other cards of the nemesis set go to the encounter discard pile.
- **Encounter deck empty:** shuffle the encounter discard pile into a new encounter deck and add an acceleration token. If both the encounter deck and its discard pile are empty at the same time, you lose.
- A discard effect that empties the encounter deck stops there; any other effect finishes with the new deck.
- Encounter cards dealt face down, cards in decks and discard piles, set-aside cards and boost cards are out of play.

## Player deck
- If your deck is empty, shuffle your discard pile into a new deck and deal yourself one face-down encounter card. If your discard pile is also empty, this happens as soon as a card enters it.
- If the deck runs out while you are drawing, keep drawing from the new deck. If it runs out while you are discarding from it, stop.
- After searching any part of a deck, shuffle that deck.
- Discard piles are open information; the order of decks cannot be changed unless a card says so.

## Status cards
A character can have only one status card of each type (two of stunned/confused with steady). Status cards take priority over triggered abilities.
- **Stunned:** the character's next attack is replaced by discarding the stun. A stunned hero or ally that attempts an attack still pays its costs, including exhausting.
- **Confused:** the character's next scheme or thwart is replaced by discarding the confuse. A confused hero or ally may attempt a thwart even with no threat to remove, and still exhausts.
- **Tough:** the next time the character would take any damage, prevent all of it and discard the tough card. A hero who defends and reduces the damage to 0 with DEF keeps the tough card.

## Keywords
- **Assault:** a basic thwart against this scheme uses ATK instead of THW (allies take their ATK consequential damage).
- **Guard:** while a guard minion is engaged with you, you cannot attack the villain.
- **Hinder X:** enters play with X extra threat on it.
- **Incite X:** when revealed, place X threat on the main scheme.
- **Linked (title):** set aside until the named card brings it in.
- **Overkill:** see Damage.
- **Patrol:** while a patrol minion is engaged with you, you cannot thwart the main scheme.
- **Permanent:** cannot leave play except by cards of its own set.
- **Piercing:** the attack discards the target's tough status cards before dealing damage.
- **Quickstrike:** after this minion engages you while you are in hero form, it attacks you.
- **Ranged:** the attack ignores retaliate.
- **Restricted:** you may control at most two restricted cards; discard down at once if you exceed it.
- **Retaliate X:** after this character is attacked (and is still in play), deal X damage to the attacker.
- **Setup:** the card starts the game in play.
- **Stalwart:** cannot be stunned or confused.
- **Steady:** stunned or confused only with two such status cards.
- **Surge:** when revealed, you are dealt one more face-down encounter card, revealed after this one finishes.
- **Team-Up (A and B):** can be played only while both named characters are friendly and in play.
- **Teamwork (trait):** after this minion enters play, if another minion with that trait is in play, it activates against you.
- **Temporary:** discarded at the end of the round.
- **Toughness:** enters play with a tough status card.
- **Uses (X type):** enters play with X counters; discarded when the last is removed.
- **Victory X:** when defeated, goes to the victory display.
- **Villainous:** this minion gets a boost card when it activates.
- **Vulnerable:** discarded (not defeated) when it becomes stunned or confused.

## Villain defeat
- When the villain's hit points reach 0, that stage is defeated and removed from the game; the next stage is revealed with its own hit points. Excess damage does not carry over. Defeating the final stage wins.
- If the next stage has the same title, attachments, status cards and non-damage tokens stay on it, and an activation it was in the middle of continues.
- Revealing a villain stage cannot be cancelled. When Revealed abilities on villain and main scheme cards cannot be cancelled.

## Other terms
- **Ready / exhausted:** cards enter play ready; using many abilities exhausts a card, and an exhausted card cannot exhaust again until readied. Abilities on exhausted cards still work if they do not need the card to exhaust.
- **Attach to:** an attached card stays until it or its host leaves play; it is discarded when its host leaves play.
- **Leaves play:** a card that leaves play loses all damage, tokens and status cards, its attachments are discarded, and it returns as a new copy with no memory.
- **Unique:** you cannot have a unique card in play if a matching unique card (same title or alias) is already in play.
- **Modifiers:** values are recalculated whenever a modifier changes; additions and subtractions come before doubling and halving; values below 0 count as 0; fractions round up; "set" overrides other modifiers.
- **Printed:** the value or text physically on the card.
- **"You":** refers to your identity when it can, otherwise to you as the player.
- **Search** lets you look at cards in an area and take one that matches; **find** looks in every game area except face-down encounter cards, the victory display and removed cards.
