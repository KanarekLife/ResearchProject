# How to play Marvel Champions solo

You are one hero against a villain run by the game. These rules follow the official Rules Reference v1.8. If a card's text disagrees with a rule here, the card wins. If something is unclear, assume the outcome that is worse for you.

## 1. The two ways the game ends
- **You win** when the villain's last stage reaches 0 hit points. The villain has several stages; defeating one reveals the next at full hit points, and leftover damage is lost.
- **You lose** when your identity reaches 0 hit points, or when threat on the main scheme reaches its target threat.

Every decision is a trade between these two clocks: damage on you and threat on the main scheme.

## 2. Your identity has two forms
Your identity card has a hero side and an alter-ego side. You start as the alter-ego.

| | Hero form | Alter-ego form |
|---|---|---|
| Basic powers | ATK: attack an enemy. THW: remove threat from a scheme. DEF: defend against an attack. | REC: heal that much damage (only if you are damaged). |
| What enemies do in the villain phase | They **attack** you. | They **scheme**: add their SCH as threat to the main scheme. |

- Using a basic power exhausts your identity, so you get one basic power per turn unless something readies you.
- You may change form once during your turn. A card that changes your form does not count against this.
- Damage, status cards, upgrades and exhausted/ready state stay when you change form.
- Cards and abilities marked "Hero" or "Alter-Ego" only work in that form. Each form has its own hand size.

## 3. A round
**Player phase (your turn).** Do any of these, in any order, as often as you can pay:
1. Play allies, supports and upgrades (they stay in play) or events (they resolve and go to your discard pile).
2. Use one basic power with your identity.
3. Exhaust each ready ally to attack or thwart.
4. Trigger Action abilities on your cards, on encounter cards in play, or from events in hand.
5. Change form once.

Then end your turn: discard any cards you want (you must discard down to your hand size if you are above it), draw up to your hand size, and ready everything you control.

**Villain phase.** The game runs it; you make choices when asked.
1. Threat is added to the main scheme: its per-round threat + 1 per acceleration icon + 1 per acceleration token.
2. The villain activates against you, then each minion engaged with you, one by one.
3. You are dealt 1 face-down encounter card + 1 per hazard icon in play.
4. Those cards (and any others dealt to you this round) are revealed and resolved one at a time.

## 4. Paying for cards
- Pay a card's cost by discarding other cards from your hand. Each discarded card gives the resources printed in its bottom-left corner. Some cards print two resources, and some cards in play have "Resource" abilities that generate one.
- Resource types: energy, mental, physical, wild. Wild counts as any type.
- Usually any types work. Some cards need specific types, and some get a bonus if you paid with a certain type; then the type matters.
- Extra resources are lost. A card cannot pay for itself.
- Every card in hand is both something you can play and fuel for other cards. Playing one card usually costs you one or more others.

## 5. When an enemy attacks you
1. The villain gets a face-down **boost card** from the encounter deck (minions do not, unless they are villainous).
2. You declare a defender:
   - **Your hero** (exhaust): the damage is reduced by your DEF; you take the rest.
   - **One of your allies** (exhaust): the ally takes all the damage. If it is defeated, the excess is lost, unless the attack has overkill, in which case the excess hits you.
   - **No defender:** you take all the damage.
   Only a ready character can defend. You declare before you see the boost card. A defender stays exhausted until you ready at the end of your next player phase, so a hero that defends cannot use a basic power on your next turn.
3. The boost card flips: each boost icon adds +1 ATK. A star in the boost area means the card has a "Boost:" ability that also resolves. Nothing else on that card is used.
4. The damage is dealt.

**Damage dealt to you = villain ATK + attachment bonuses + boost icons − DEF (if your hero defends).** A minion's attack is its ATK with no boost.

## 6. When an enemy schemes
The villain gets a boost card; its boost icons add to its SCH, then that much threat goes on the main scheme. Minions add their SCH with no boost.

**Threat added in a villain phase while you are an alter-ego = per-round threat + acceleration icons + acceleration tokens + villain SCH + boost icons + each engaged minion's SCH + any threat from revealed encounter cards.**

While you are a hero, the first three terms still apply, but the villain and minions attack instead of scheming.

## 7. Allies
- You may control at most 3 allies. Playing a 4th forces you to discard one immediately.
- An ally attacks or thwarts by exhausting. Afterwards it takes **consequential damage** equal to the damage icons under the ATK or THW it used.
- An ally can defend for you (see section 5). Allies are defeated at 0 hit points and go to your discard pile.

## 8. Encounter cards
What a revealed card does depends on its type:
- **Minion:** engages you. Each villain phase it attacks you (hero) or schemes (alter-ego). Defeat it by reducing its hit points to 0.
- **Treachery:** a one-time effect, then discarded. Many say "Hero:" or "Alter-Ego:" and only do that part in that form.
- **Side scheme:** enters play with threat on it. Thwart it to 0 to discard it. Many have icons:
  - **Acceleration:** +1 threat on the main scheme every villain phase.
  - **Hazard:** +1 encounter card dealt every villain phase.
  - **Crisis:** while it is in play, your cards cannot remove threat from the main scheme.
- **Attachment:** usually attaches to the villain and changes its stats or attacks. Some can be removed by an action printed on them.
- **Obligation:** a problem tied to your alter-ego. You resolve it by its own instructions.
- **Environment:** a lasting rule change.
- **Surge:** you are dealt one more encounter card.
- **Nemesis:** some cards bring in your nemesis: a nemesis minion engaged with you and a nemesis side scheme.

If the encounter deck runs out, its discard pile is shuffled into a new deck and an acceleration token is added permanently.

## 9. Statuses
- **Stunned:** the character's next attack does nothing except remove the stun. A stunned hero still exhausts when attempting it.
- **Confused:** the character's next thwart or scheme does nothing except remove the confuse.
- **Tough:** the next damage the character would take is prevented completely, and the tough card is removed. If your hero defends and DEF reduces the damage to 0, the tough card stays.

A character can have at most one of each.

## 10. Keywords you may meet
- **Guard:** while a guard minion is engaged with you, you cannot attack the villain.
- **Patrol:** while a patrol minion is engaged with you, you cannot thwart the main scheme.
- **Quickstrike:** the minion attacks you as soon as it engages you, if you are a hero.
- **Overkill:** excess damage after defeating an ally goes to you (or, on your attacks, after defeating a minion goes to the villain).
- **Retaliate X:** after you attack this character, you take X damage.
- **Piercing:** the attack removes tough statuses before dealing damage.
- **Ranged:** the attack ignores retaliate.
- **Toughness:** enters play with a tough status.
- **Stalwart:** cannot be stunned or confused. **Steady:** needs two stuns or two confuses.
- **Villainous:** the minion gets a boost card like the villain.
- **Incite X:** when revealed, adds X threat to the main scheme. **Hinder X:** enters play with X extra threat.
- **Assault:** a basic thwart against this scheme uses ATK instead of THW.
- **Teamwork:** the minion activates when it arrives if a minion with the same trait is in play.
- **Temporary:** discarded at the end of the round. **Uses (X):** discarded when its X counters are spent.
- **Restricted:** you may control at most two restricted cards.
- **Permanent:** cannot leave play except through cards of its own set.

## 11. Card abilities and timing
- **Action:** you choose to use it during your turn.
- **Interrupt:** used just before something happens, and can change or stop it.
- **Response:** used just after something happens.
- **Forced**, **When Revealed**, **When Defeated** and **Boost** abilities happen automatically.
- An ability that needs a target cannot be used without a valid target.
- A "Hero Action (attack)" or "(thwart)" ability counts as your hero attacking or thwarting, but it does not exhaust your hero unless it says so.
- "Then": the second part happens only if the first part fully happened.
- Limits such as "Limit once per round" apply to each copy of the card.

## 12. Running out of cards
- **Your deck:** when it is empty, your discard pile is shuffled into a new deck and you are dealt one encounter card.
- **Encounter deck:** see section 8. If the encounter deck and its discard pile are both empty, you lose.

## 13. Other rules that matter
- Damage beyond what defeats a minion, ally or villain stage is lost (except with overkill).
- Healing cannot take you above your maximum hit points.
- A card that leaves play loses its damage, statuses and tokens, and the cards attached to it are discarded.
- Unique cards: you cannot have two cards with the same unique title in play.
- Discard piles are open information. The order of decks is secret and cannot be changed unless a card says so.
