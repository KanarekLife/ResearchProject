package openaicompatible

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func statusError(resp *http.Response, raw []byte, useMaxCompletion bool) error {
	if resp.StatusCode == http.StatusBadRequest && !useMaxCompletion && strings.Contains(string(raw), fieldMaxCompletionTokens) {
		return errTokenParam
	}
	return &httpError{status: resp.StatusCode, retryAfter: retryAfter(resp), body: string(raw)}
}

func retryable(err error) bool {
	var he *httpError
	if errors.As(err, &he) {
		return he.status == http.StatusTooManyRequests || he.status >= 500
	}
	return true
}

// Backoff bounds for retrying a transient failure. backoffBase is a variable
// only so tests need not wait.
var backoffBase = 2 * time.Second

const backoffMax = 30 * time.Second

// backoff sleeps before a retry, honoring a Retry-After hint and the context.
func (c *Client) backoff(ctx context.Context, attempt int, cause error) {
	delay := min(backoffBase*time.Duration(1<<(attempt-1)), backoffMax)
	var he *httpError
	if errors.As(cause, &he) && he.retryAfter > 0 {
		delay = he.retryAfter
	}
	t := time.NewTimer(delay)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

func retryAfter(resp *http.Response) time.Duration {
	if s := resp.Header.Get(headerRetryAfter); s != "" {
		if secs, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
			return time.Duration(secs) * time.Second
		}
	}
	return 0
}
