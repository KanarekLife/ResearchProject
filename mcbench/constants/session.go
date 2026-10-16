package constants

// Decision kinds.
const (
	KindMulligan = "mulligan"
	KindTurn     = "turn"
	KindDefend   = "defend"
	KindWindow   = "window"
	KindChoice   = "choice"
	KindDiscard  = "discard"
	KindPayment  = "payment"
)

// Game log lines for a choice. The engine writes EventChose for every applied
// option; the session rewrites it to EventAutoResolved for a forced decision.
const (
	EventChose        = "Player chose: "
	EventAutoResolved = "Auto-resolved (only legal option): "
)

// Option keys and key prefixes. Option ids are 1..n in the order shown; keys
// are stable engine labels used in records and by scripted players.
const (
	KeyPass            = "pass"
	KeyEndTurn         = "end_turn"
	KeyMulliganKeep    = "mulligan:keep"
	KeyMulliganDiscard = "mulligan:discard"
	KeyDiscardDone     = "discard:done"
	KeyCancel          = "cancel"

	PrefixMulligan   = "mulligan:"
	PrefixPlay       = "play:"
	PrefixAttack     = "attack:"
	PrefixThwart     = "thwart:"
	PrefixDefend     = "defend:"
	PrefixChangeForm = "change_form:"
	PrefixRecover    = "recover:"
	PrefixDiscard    = "discard:"
	PrefixUse        = "use:"
	PrefixEffect     = "effect:"
	PrefixResource   = "res:"
	PrefixPay        = "pay:"
)

// Separators inside an option key.
const (
	SepTarget  = ">"     // "play:Swinging Web Kick>Rhino"
	SepPayment = "|pay:" // "...>Rhino|pay:Genius+Energy"
	SepLabel   = "#"     // "Name#ID" when names repeat in play
	SepPayJoin = "+"     // between payment sources
)
