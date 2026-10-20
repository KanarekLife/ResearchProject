package systemone

// DefaultBaseURL is TypeSafe's own endpoint; Codiv serves the same API at
// https://api.codiv.ai/v1.
const DefaultBaseURL = "https://api.typesafe.ai/v1"

// Wire-format names of the System One API.
const (
	pathSystemOne = "/systemone"

	headerAuthorization = "Authorization"
	bearerPrefix        = "Bearer "

	questionTypeChoice = "choice"
	// questionName names the one question every request asks.
	questionName = "decision"
)
