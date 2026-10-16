# How this game is run

The game is played by a rules engine. You never move cards yourself: at every point where the rules give you a choice, the engine stops and lists every legal option. You pick one by its id. Everything else (shuffling, threat, enemy activations, encounter cards, forced and mandatory abilities) is resolved for you.

## Tools
- **get_decision**: the decision you must make now (`kind`, `prompt`, numbered `options`), or the final status once the game is over.
- **choose_option**: makes the current decision. Pass the `option_id` of a listed option and a short `reasoning`. The result lists the events that followed and the **next decision**, so you do not need to call get_decision again after it.
- **get_state**: the visible board: round and phase; your hero (form, HP, ATK/THW/DEF/REC, hand size, exhausted, statuses, whether you already changed form this turn, the other form's name and ability); your hand with costs and resource icons; your cards in play; deck size and discard pile; face-down encounter cards dealt to you; the villain (stage, HP, ATK, SCH, keywords, statuses, attachments) and its later stages; the main scheme (threat, the threshold at which you lose, threat per round); side schemes with their icons; minions engaged with you; acceleration tokens; encounter deck size and discard pile. It never shows the order of a deck or face-down cards.
- **get_card**: the type, cost, stats and text of a card by name. Use it whenever you are unsure what a card does.
- **get_log**: the last lines of the game log (`last`, default 20).

## Decisions
Each decision has a kind:
- **mulligan**: at the start. Each option discards one card from your opening hand; you are asked again after each discard. "Keep this hand" ends the mulligan and draws you back up to your hand size.
- **turn**: your main window in the player phase. The options are every action you can take right now: playing a card, a basic power, an ally's attack or thwart, a card's action, changing form, or ending your turn. After each action you get a new turn decision with the options that are legal then.
- **discard**: after you end your turn. Each option discards one card from your hand. If you are above your hand size you must discard until you are not; otherwise "Done discarding" draws you up to your hand size and the villain phase follows.
- **defend**: an enemy is attacking you. Choose your hero, a ready ally, or no defender. You choose before the villain's boost card is flipped.
- **window**: a moment where a card in your hand can interrupt or respond (for example when threat is about to be placed, or when you are about to be attacked). "Do not respond" (`pass`) lets the event happen. You can play at most one card in each window.
- **choice**: a card or rule asks you to pick something, such as which enemy or scheme an effect targets, or which ally to discard when you are over the ally limit.

## Options
- Each option is one complete action. Its text says what it does, including the target and, for cards with a cost, exactly which cards you discard or use to pay. The same card can appear several times with different targets or different payments; choose the payment that keeps the cards you still want.
- Each option has a stable `key`, for example `play:Card Name>Target|pay:Card A+Card B`, `attack:Hero>Enemy`, `thwart:Hero>Scheme`, `defend:none`, `change_form:Name`, `end_turn`. When two cards in play share a name, they are labelled `Name#N`.
- Option ids are numbered 1..n for that decision only, and the order of options is shuffled. The position of an option means nothing; read its text.
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
- You get a limited number of replies for each decision (five in the default configuration). Each reply can call several tools, but once one choose_option succeeds, the rest of that reply's calls are rejected. A reply without any tool call uses up one of your replies.
- An unknown option id returns an error and counts as an invalid choice. If you run out of replies without a valid choice, the game ends as `agent_error` with a score of 0.
- The whole conversation is kept, so earlier tool results stay available to you; call get_state when you need the current board.
