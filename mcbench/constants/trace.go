package constants

// JSON keys and type values of the per-game trace written by the session,
// the model player and the benchmark harness.
const (
	TraceKeyType      = "type"
	TraceKeyTime      = "time"
	TraceKeyDecision  = "decision"
	TraceKeyChoice    = "choice"
	TraceKeyText      = "text"
	TraceKeyEvents    = "events"
	TraceKeyStatus    = "status"
	TraceKeyMessage   = "message"
	TraceKeyError     = "error"
	TraceKeyRounds    = "rounds"
	TraceKeyCriteria  = "criteria"
	TraceKeyScore     = "score"
	TraceKeyUsage     = "usage"
	TraceKeyScenario  = "scenario"
	TraceKeySeed      = "seed"
	TraceKeySample    = "sample"
	TraceKeyPlayer    = "player"
	TraceKeyInstr     = "instructions"
	TraceKeyInputTok  = "prompt_tokens"
	TraceKeyOutputTok = "completion_tokens"
	TraceKeyCachedTok = "cached_tokens"
	TraceKeyTruncated = "truncated"
)

// Trace entry types.
const (
	TraceStart   = "start"
	TraceEnd     = "end"
	TraceChoice  = "choice"
	TraceMessage = "message"
)
