# How this game is run

The game is played by a rules engine. You never move cards yourself: at every point where the rules give you a choice, the engine stops and lists every legal option. You pick one by its id. Everything else (shuffling, threat, enemy activations, encounter cards, forced and mandatory abilities) is resolved for you.

## Tools
- **get_decision**: the decision you must make now (`kind`, `prompt`, numbered `options`), or the final status once the game is over.
- **choose_option**: makes the current decision. Pass the `option_id` of a listed option and a short `reasoning`. The result lists the events that followed and the **next decision**, so you do not need to call get_decision again after it.
- **get_state**: the visible board: round and phase; your hero (form, HP, ATK/THW/DEF/REC, hand size, exhausted, statuses, whether you already changed form this turn, the other form's name and ability); your hand with costs and resource icons; your cards in play; deck size and discard pile; face-down encounter cards dealt to you; the villain (stage, HP, ATK, SCH, keywords, statuses, attachments) and its later stages; the main scheme (threat, the threshold at which you lose, threat per round); side schemes with their icons; minions engaged with you; acceleration tokens; encounter deck size and discard pile. It never shows the order of a deck or face-down cards.
- **get_card**: the type, cost, stats and text of a card by name or code. Close spellings and partial names are accepted; when several cards match (such as the stages of a villain), it lists them with their codes and you call it again with the code you want. Use it whenever you are unsure what a card does.
- **get_log**: the last lines of the game log (`last`, default 20).

## Decisions
Each decision has a kind:
- **mulligan**: at the start. Each option discards one card from your opening hand; you are asked again after each discard. "Keep this hand" ends the mulligan and draws you back up to your hand size.
- **turn**: your main window in the player phase. The options are every action you can take right now: playing a card, a basic power, an ally's attack or thwart, a card's action, changing form, or ending your turn. After each action you get a new turn decision with the options that are legal then.
- **discard**: after you end your turn. Each option discards one card from your hand. If you are above your hand size you must discard until you are not; otherwise "Done discarding" draws you up to your hand size and the villain phase follows.
- **defend**: an enemy is attacking you. Choose your hero, a ready ally, or no defender. You choose before the villain's boost card is flipped.
- **window**: a moment where a card in your hand can interrupt or respond (for example when threat is about to be placed, or when you are about to be attacked). "Do not respond" (`pass`) lets the event happen. You can play at most one card in each window.
- **choice**: a card or rule asks you to pick something, such as which enemy or scheme an effect targets, or which ally to discard when you are over the ally limit.
- **payment**: you chose a card or ability that can be paid for in more than one way. Each option is one way to pay (`pay:Card A+Card B`); `cancel` goes back to the previous decision without paying or playing anything.

## Options
- Each option is one action. Its text says what it does, including the target. A card that can be paid for in only one way (or is free) shows that payment in its text and is played at once. A card with several ways to pay shows no payment and is followed by a **payment** decision; choose the payment that keeps the cards you still want, or cancel. The same card appears once per target.
- Each option has a stable `key`, for example `play:Card Name>Target` (with `|pay:Card A+Card B` when the payment is fixed), `pay:Card A+Card B`, `attack:Hero>Enemy`, `thwart:Hero>Scheme`, `defend:none`, `change_form:Name`, `cancel`, `end_turn`. When two cards in play share a name, they are labelled `Name#N`.
- Option ids are numbered 1..n for that decision only. Options are grouped by kind: on your turn, card plays first, then your hero's basic powers, ally attacks and thwarts, card actions, changing form, and ending your turn last. Read the text of an option, not its position.
- Only listed options are legal. If an action is not listed, you cannot do it now (for example, it cannot be paid for, it has no valid target, or your hero is exhausted).

## Things the engine does automatically
- Mandatory abilities: forced interrupts and responses, When Revealed, When Defeated, Boost abilities and keywords.
- A few optional responses that can only help you are triggered without asking.
- Drawing up to your hand size and readying at the end of the player phase.

## Status
The status in each result is one of:
- `awaiting_decision`: the game is waiting for your choice.
- `won` or `lost`: the game is over.
- `round_limit`: the game stopped because the maximum number of rounds was reached. This is not a win.
- `decision_limit` or `engine_error`: the game stopped for a technical reason.

## Limits on your replies
- Looking things up is free: you may call get_state, get_card, get_log and get_decision as often as you need before choosing.
- A reply fails when it has no tool call or one of its calls returns an error (for example an unknown option id, which also counts as an invalid choice, or a card name that matches nothing). After five failed replies for one decision (in the default configuration) the game ends as `agent_error` with a score of 0. A decision that takes 100 replies ends the same way.
- Call choose_option once per reply. Once it succeeds, the rest of that reply's calls are rejected, because the options change after every choice.
- The whole conversation is kept, so earlier tool results stay available to you; call get_state when you need the current board.
