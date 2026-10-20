package systemone

type request struct {
	Model     string              `json:"model"`
	State     any                 `json:"state"`
	Questions map[string]question `json:"questions"`
}

type question struct {
	Type         string            `json:"type"`
	Instructions string            `json:"instructions"`
	Criteria     map[string]string `json:"criteria"`
}

type response struct {
	Answers map[string]struct {
		Choice        string             `json:"choice"`
		Confidence    float64            `json:"confidence"`
		Probabilities map[string]float64 `json:"probabilities"`
	} `json:"answers"`
	Usage struct {
		InputTokens int64 `json:"input_tokens"`
	} `json:"usage"`
}
