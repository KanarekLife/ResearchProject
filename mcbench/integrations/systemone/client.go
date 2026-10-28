// Package systemone is a client for the System One API (typesafe.ai, also
// served by codiv.ai): a classifier that answers a multiple-choice question
// about a piece of state with a probability for every option. It is not a chat
// model, so it does not implement inference.Client.
package systemone

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"time"

	"mcbench/integrations/inference"
)

// Config configures a Client.
type Config struct {
	BaseURL string
	APIKey  string
	Model   string
	Retries int
	// RateLimitWait is how long rate limits are waited out in all; 0 retries
	// them as transient failures.
	RateLimitWait time.Duration
	Headers       map[string]string
}

// Option is one answer the model may pick. Name is the key the API answers
// with; Description tells the model what it means.
type Option struct {
	Name        string
	Description string
}

// Answer is the model's pick: the top option, its probability and the whole
// distribution.
type Answer struct {
	Choice        string
	Confidence    float64
	Probabilities map[string]float64
	InputTokens   int64
}

// Client asks System One choice questions.
type Client struct{ cfg Config }

func New(cfg Config) *Client { return &Client{cfg: cfg} }

// Choose asks which option best answers instructions, given state (any value
// that marshals to JSON).
func (c *Client) Choose(ctx context.Context, state any, instructions string, options []Option) (Answer, error) {
	criteria := make(map[string]string, len(options))
	for _, o := range options {
		criteria[o.Name] = o.Description
	}
	body, err := json.Marshal(request{
		Model: c.cfg.Model, State: state,
		Questions: map[string]question{questionName: {Type: questionTypeChoice, Instructions: instructions, Criteria: criteria}},
	})
	if err != nil {
		return Answer{}, err
	}
	url := strings.TrimRight(c.cfg.BaseURL, "/") + pathSystemOne
	headers := map[string]string{}
	if c.cfg.APIKey != "" {
		headers[headerAuthorization] = bearerPrefix + c.cfg.APIKey
	}
	maps.Copy(headers, inference.ExpandHeaders(ctx, c.cfg.Headers))
	var ans Answer
	// Retry is shared with the chat clients; the answer leaves through ans.
	_, err = inference.Retry(ctx, c.cfg.Retries, c.cfg.RateLimitWait, func() (inference.Reply, error) {
		raw, err := inference.PostJSON(ctx, url, headers, body)
		if err != nil {
			return inference.Reply{}, err
		}
		ans, err = decode(raw)
		return inference.Reply{InputTokens: ans.InputTokens}, err
	})
	return ans, err
}

func decode(raw []byte) (Answer, error) {
	var r response
	if err := json.Unmarshal(raw, &r); err != nil {
		return Answer{}, fmt.Errorf("decoding system one reply: %w", err)
	}
	a, ok := r.Answers[questionName]
	if !ok || a.Choice == "" {
		return Answer{}, fmt.Errorf("system one reply has no %q choice", questionName)
	}
	return Answer{Choice: a.Choice, Confidence: a.Confidence, Probabilities: a.Probabilities, InputTokens: r.Usage.InputTokens}, nil
}
