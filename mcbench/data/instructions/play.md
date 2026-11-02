# How to play to win (solo)

This is a standing order for every decision, not advice. It works for any hero against any villain. Follow the steps in order. Where your instinct says to save cards or protect your hit points, this procedure wins: holding back is the most common way to lose. Where a card's text disagrees with this procedure, the card wins.

## The two clocks

You lose when your hit points reach 0, or when threat on the final main scheme stage reaches its target. Every round the villain survives brings another villain phase, more threat and more encounter cards, so the default is to attack. But pushing on one clock while ignoring the other also loses: count both every turn.

Read the numbers with get_state:

- **T**: threat on the main scheme. **Target**: its target threat.
- **Hero threat** (the next villain phase, if you end your turn as a hero) = per-round threat + acceleration icons + acceleration tokens + **2** (an encounter card that adds threat).
- **Alter-ego threat** = hero threat + the villain's SCH + each engaged minion's SCH + **2** (a high boost).
- **Incoming damage** (if you end as a hero) = the villain's ATK + attachment bonuses + **2** (a high boost) + each engaged minion's ATK.
- **Your damage**: the hero's basic attack if ready, each ready ally's ATK, attack events you can pay for, and card actions that deal damage.
- **Rounds to win** = the villain's hit points left, on this stage and every later stage, ÷ the damage you deal per round.
- **Rounds on the scheme** = (Target − T) ÷ hero threat. **Rounds on your HP** = your HP ÷ incoming damage.

Boost cards usually carry 0, 1 or 2 icons, rarely 3, so 2 is the planning value.

## The clock check

Before your first action of every turn, write one line in your reasoning:

`Clocks: T x/Target, +h hero / +a alter-ego; HP y vs incoming d; win in w rounds, scheme lasts s, HP lasts p.`

Then decide which clock this turn must serve:

1. **Lethal.** Your damage ≥ the villain's remaining HP on its final stage: spend everything on the villain.
2. **Scheme in danger.** T + hero threat ≥ Target: remove threat until it is not, before anything else.
3. **HP in danger.** Incoming damage ≥ your HP: plan to defend, heal, prevent the damage or end the turn in alter-ego (if the scheme can take alter-ego threat).
4. **Scheme too slow.** s < w: the scheme runs out before you can win. Keep attacking, but each turn also remove enough threat (allies and events first) that s stays above w.
5. **HP too slow.** p < w: you will be defeated before you win. Plan the turn when you recover in alter-ego or defend with an ally, before HP gets below incoming damage.
6. **Otherwise attack.** Both clocks are safe: every action goes into damage and building your board.

Do the check again before ending the turn.

## Your turn (`turn` decisions)

Repeat after every action, taking the first step that applies.

1. Apply the clock check.
2. **Draw and resources first.** Use card actions that draw cards or give resources before you choose what to play.
3. **Alter-ego first.** In alter-ego (always in round 1), use the alter-ego's ability and alter-ego-only cards before you change form.
4. **Build the board.** Play every ally, support and upgrade you can pay for. Pay with events you cannot use in the next two turns, expensive cards and duplicates.
5. **Change to hero** if you are in alter-ego and have not changed form, unless the end-of-turn check says to end in alter-ego.
6. **Dangerous enemies.** Defeat a minion with guard, patrol, or ATK or SCH of 2 or more, with the smallest attack that is enough. Clear a side scheme with acceleration, hazard or crisis, completely if you can, with an ally or an event. Crisis stops your cards removing threat from the main scheme, so clear it before the main scheme needs thwarting.
7. **Use everything.** Each ready ally and the hero's basic power: attack the villain, or thwart if the clock check said the scheme needs it. Use every card action. An action not used this turn is lost, not saved.
8. **End the turn** only when no ready ally, basic power, card action or payable card is left. Keep a card only if it responds to something (an interrupt or a response) or you cannot pay for it.

**End in alter-ego** only if all three hold: T + alter-ego threat < Target; the villain is not stunned (a stun is wasted while it schemes); and you need it (incoming damage could defeat you or leave you below the next attack, or you will recover next turn). Otherwise end in hero form. You can change form only once per turn.

## Read the game you are given

Heroes, villains and encounter sets differ. Learn them instead of assuming:

- **At the start**, read your identity with get_state (both forms, their abilities and hand sizes) and look up each villain stage and each main scheme stage with get_card. Note what changes between stages.
- **A card you do not know**: call get_card before you play it, pay with it or decide against it. Its text decides how it is used.
- **Villain stages.** Defeating a stage reveals the next at full hit points; damage beyond a stage's hit points is lost, so finish a stage with the smallest effect that is enough.
- **Main scheme stages.** If the villain has more than one, threat reaching the target of an earlier stage advances the scheme to the next stage (excess threat is lost, and the new stage may add threat or other effects when revealed). Only the final stage loses the game. Treat every advance as a setback: its effects can hurt.
- **Keywords** on minions, schemes and the villain (guard, patrol, retaliate, toughness, quickstrike, overkill, surge and others) are explained in the rules. Account for them in the clock check: retaliate costs you HP, toughness costs one extra hit, surge and hazard bring extra encounter cards.

## Defend (`defend` decisions)

You choose before the boost is known; use incoming damage with a high boost.

1. A card in hand prevents this damage or cancels this attack and you can pay for it: choose no defender, and play it in the window that follows.
2. An ally that will die to its own consequential damage soon, or will be discarded anyway, defends.
3. The attack could defeat you, or leave you with less HP than the next attack with a high boost: defend, with the ally whose next attack or thwart matters least, otherwise with the hero.
4. Otherwise choose **no defender**. A hero that defends stays exhausted through your next turn and loses its basic power; that is usually worth more than the few HP its DEF saves.

## Windows, payments, choices, discards

- `window`: play a card that cancels a treachery, attack or surge that would hurt you, or prevents damage that would leave you below the next attack, if paying for it leaves your key cards for next turn. Where a card trades damage for threat (or the reverse), take the one whose clock has more rounds left. Otherwise `pass`.
- `payment`: pay with the cards you would discard at the end of the turn anyway, and with resource abilities of cards in play before cards from your hand. Never pay with the card you plan to play next. Where a card's effect depends on the resource type paid, pay with that type.
- `choice`: pick the target that ends a danger soonest: the minion that dies, the scheme that empties, the ally that is about to be discarded anyway.
- `discard`: above your hand size, discard what you are least likely to play. Otherwise discard cards you cannot use in the next two turns, but not when your deck has fewer cards than you would draw: each time your deck runs out you are dealt an extra encounter card.

## Mulligan

Keep allies, supports and upgrades that cost 3 or less, and cards that give resources or draw cards. Discard expensive cards, duplicates and events that answer a later situation. Discarding 2 to 4 cards is normal.

## Checks that override everything

- Never end a turn in alter-ego when T + alter-ego threat ≥ Target.
- Never end a turn with a ready ally or an unused basic power because you were saving it.
- Never defend with the hero when step 4 of `defend` applies.
- Never end a turn without the clock check.
