package constants

// Decision kinds.
const (
	KindMulligan = "mulligan"
	KindTurn     = "turn"
	KindDefend   = "defend"
	KindWindow   = "window"
	KindChoice   = "choice"
	KindDiscard  = "discard"
)

// Option keys and key prefixes. Option ids are 1..n in the order shown; keys
// are stable engine labels used in records and by scripted players.
const (
	KeyPass            = "pass"
	KeyEndTurn         = "end_turn"
	KeyMulliganKeep    = "mulligan:keep"
	KeyMulliganDiscard = "mulligan:discard"
	KeyDiscardDone     = "discard:done"

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
)

// Separators inside an option key.
const (
	SepTarget  = ">"     // "play:Swinging Web Kick>Rhino"
	SepPayment = "|pay:" // "...>Rhino|pay:Genius+Energy"
	SepLabel   = "#"     // "Name#ID" when names repeat in play
	SepPayJoin = "+"     // between payment sources
)
