package inference

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// requestTimeout bounds one HTTP request. Reasoning models can be slow.
const requestTimeout = 10 * time.Minute

var httpClient = &http.Client{Timeout: requestTimeout}

// BackoffBase is the first retry delay, doubled per attempt up to backoffMax.
// It is a variable only so tests need not wait.
var BackoffBase = 2 * time.Second

const backoffMax = 30 * time.Second

// HTTPError is a non-200 response.
type HTTPError struct {
	Status     int
	RetryAfter time.Duration
	Body       string
}

func (e *HTTPError) Error() string { return fmt.Sprintf("API HTTP %d: %.300s", e.Status, e.Body) }

// PostJSON posts body to url with the given headers and returns the response
// body of a 200, or an *HTTPError.
func PostJSON(ctx context.Context, url string, headers map[string]string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		he := &HTTPError{Status: resp.StatusCode, Body: string(raw)}
		if secs, err := strconv.Atoi(strings.TrimSpace(resp.Header.Get("Retry-After"))); err == nil {
			he.RetryAfter = time.Duration(secs) * time.Second
		}
		return nil, he
	}
	return raw, nil
}

// Retry runs do, retrying rate limits, server errors and network failures up
// to retries more times, with exponential backoff that honors Retry-After.
func Retry(ctx context.Context, retries int, do func() (Reply, error)) (Reply, error) {
	attempts := max(1, retries+1)
	var last error
	for attempt := 1; attempt <= attempts; attempt++ {
		if attempt > 1 {
			sleep(ctx, attempt-1, last)
		}
		reply, err := do()
		if err == nil {
			return reply, nil
		}
		last = err
		slog.Warn("chat request failed", "attempt", attempt, "of", attempts, "error", err)
		if !retryable(err) {
			return Reply{}, err
		}
	}
	return Reply{}, fmt.Errorf("chat request failed after %d attempts: %w", attempts, last)
}

func retryable(err error) bool {
	var he *HTTPError
	if errors.As(err, &he) {
		return he.Status == http.StatusTooManyRequests || he.Status >= 500
	}
	return true
}

func sleep(ctx context.Context, retry int, cause error) {
	delay := min(BackoffBase*time.Duration(1<<(retry-1)), backoffMax)
	var he *HTTPError
	if errors.As(cause, &he) && he.RetryAfter > 0 {
		delay = he.RetryAfter
	}
	t := time.NewTimer(delay)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
