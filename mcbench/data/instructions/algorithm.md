# Play algorithm (solo)

Follow this algorithm at every decision. Work through the steps in order and take the first rule that applies. It is a procedure, not advice: do not skip a step because an action looks costly. Holding back cards, actions or the hero's basic power is the most common way to lose.

## The numbers you need

Read them with get_state at the start of each turn and again whenever the board changes.

- **T**: threat on the main scheme. **Target**: the threat at which you lose.
- **Hero threat**: the threat the next villain phase adds if you end the turn in hero form = per-round threat + acceleration icons + acceleration tokens + **2** (an encounter card that adds threat).
- **Alter-ego threat**: hero threat + the villain's SCH + each engaged minion's SCH + **2** (a high boost).
- **Incoming damage**: if you end in hero form = the villain's ATK + **2** (a high boost) + each engaged minion's ATK.
- **Your damage**: everything you can deal this turn: the hero's basic attack (if ready), each ready ally's ATK, attack events you can pay for, and card actions that deal damage.

The scheme is **in danger** when T + hero threat ≥ Target.

## Mulligan

1. Keep allies, supports and upgrades that cost 3 or less, and cards that give extra resources or draw cards.
2. Discard events that answer a later situation (removing threat, big attacks, healing), expensive cards, and duplicates.
3. Discard 2 to 4 cards. Keeping all six is almost always wrong.

## Your turn (`turn` decisions)

Repeat these steps after every action until step 9 ends the turn.

1. **Lethal.** IF your damage ≥ the villain's remaining HP on its final stage, THEN spend everything on the villain now: change to hero if needed, use every ally, event and the basic attack. Remove a tough status with the smallest hit first.
2. **Scheme in danger.** IF the scheme is in danger, THEN remove threat until it is not: first with ally thwarts, thwart events and card actions, and only then with the hero's basic thwart. IF a crisis side scheme is in play, clear it first, because you cannot remove threat from the main scheme while it is there.
3. **Draw and resources first.** IF a card in play has an action that draws cards or gives resources, THEN use it now, before you choose what to play.
4. **Alter-ego abilities.** IF you are in alter-ego (always true in round 1), THEN use the alter-ego's ability and any card that works only in alter-ego before you change form.
5. **Build the board.** IF you can pay for an ally, support or upgrade, THEN play it. Pay with events you cannot use in the next two turns, expensive cards and duplicates. Keep a card that prevents damage or cancels an encounter card if you can pay without it.
6. **Change to hero.** IF you are in alter-ego and have not changed form this turn, THEN change to hero now, unless step 9 says this turn should end in alter-ego. In round 1 this is always the move: set up as the alter-ego, then flip and attack.
7. **Remove the dangerous enemies.** IF a minion with guard, patrol, or ATK or SCH ≥ 2 is engaged with you, THEN defeat it with the smallest set of attacks that is enough. IF a side scheme with acceleration, hazard or crisis is in play, THEN thwart it with an ally or an event, and clear it completely if you can.
8. **Attack.** For each ready ally: attack the villain. Then the hero: basic attack on the villain. Thwart with an ally or the hero instead only when T + hero threat ≥ Target − 1. An attack event goes on a minion it defeats or the villain stage it finishes; otherwise play it on the villain when nothing else is left to do. IF the villain's current stage has less HP left than the damage you are about to deal, THEN use the smallest attack that defeats it and save the rest for the next stage.
9. **End the turn.** Go through this list before choosing `end_turn`:
   - IF a ready ally, a ready hero with a basic power, a card action, or a card you can pay for is still unused, THEN go back to step 5 and use it. A card stays in hand only if it responds to something (an interrupt or a response) or you cannot pay for it.
   - Choose the form to end in. End in **alter-ego** only if all three are true: (a) T + alter-ego threat < Target; (b) the villain is not stunned; (c) you need it: incoming damage could defeat you or leave you unable to survive the next attack, or you will heal or recover next turn. Otherwise end in **hero** form.
   - IF you are in the wrong form and have not changed form this turn, THEN change form now. IF you already changed form, end the turn as you are.

## Discard (`discard` decisions)

1. IF you are above your hand size, THEN discard the cards you are least likely to play.
2. Otherwise discard each card you cannot use in the next two turns (an event for a situation that is not coming, a duplicate upgrade, an ally you cannot pay for soon), then choose "Done discarding". Keep cards that prevent damage or cancel encounter cards.
3. Do not discard to draw IF your deck has fewer cards than you would draw: each time your deck runs out you are dealt an extra encounter card.

## Villain phase: defend (`defend` decisions)

You choose before the boost is known, so use incoming damage with a high boost (ATK + 2).

1. IF you hold a card that prevents this damage or cancels this attack and can pay for it, THEN choose no defender and play that card in the window that follows.
2. IF the attack could defeat you, or leave you with less HP than the next attack with a high boost, THEN defend: with an ally that will be defeated or discarded soon anyway, otherwise with the ally whose next attack or thwart matters least, otherwise with the hero.
3. IF an ally will die to the attack's consequential damage, or is about to be discarded anyway, THEN defend with that ally.
4. Otherwise choose **no defender** and take the damage. The hero then stays ready for its basic power next turn. Defending with the hero costs next turn's basic power, which is usually worth more than the few HP its DEF saves.

## Windows (`window` decisions)

1. IF a card in hand cancels a treachery, an attack or a surge that would hurt you, and the cost leaves your key cards for next turn, THEN play it.
2. IF a card prevents damage that would leave you below the next attack's damage, THEN play it.
3. IF a card lets you take damage instead of threat (or the reverse), THEN take whichever does not bring you closer to losing: damage while your HP stays above incoming damage, threat while T stays well below Target.
4. Otherwise `pass` and keep the card.

## Payments and choices

- `payment`: pay with the cards you would discard at the end of the turn anyway. Never pay with the card you planned to play next. Use resources from abilities of cards in play before cards from your hand.
- `choice`: pick the target that ends a danger soonest: the minion that dies, the scheme that empties, the ally that is about to be discarded anyway.

## Checks that override everything

- Never end a turn in alter-ego when T + alter-ego threat ≥ Target.
- Never end a turn with a ready ally or an unused basic power because you were "saving" it: an action not used this turn is lost.
- Never defend with the hero when step 4 of `defend` applies.
