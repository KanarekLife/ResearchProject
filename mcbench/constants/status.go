package constants

// Session statuses. The harness adds AgentError; the game only reports the
// others.
const (
	AwaitingDecision = "awaiting_decision"
	Won              = "won"
	Lost             = "lost"
	RoundLimit       = "round_limit"
	DecisionLimit    = "decision_limit"
	EngineError      = "engine_error"
	AgentError       = "agent_error"
)

// Card statuses, as shown in a view.
const (
	StatusStunned  = "stunned"
	StatusConfused = "confused"
	StatusTough    = "tough"
)

// Enemy keywords, as shown in a view.
const (
	KeywordGuard       = "guard"
	KeywordQuickstrike = "quickstrike"
)

// Scheme icons, as shown in a view.
const (
	IconCrisis       = "crisis"
	IconHazard       = "hazard"
	IconAcceleration = "acceleration"
)
