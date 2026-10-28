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

const headerRetryAfter = "Retry-After"

// RateLimitBackoffBase is the first wait on a rate limit without Retry-After,
// doubled per wait up to rateLimitBackoffMax. It is a variable only so tests
// need not wait.
var RateLimitBackoffBase = 30 * time.Second

const rateLimitBackoffMax = 10 * time.Minute

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
		return nil, NewHTTPError(resp, raw)
	}
	return raw, nil
}

// NewHTTPError is the error for a non-200 resp whose body is raw.
func NewHTTPError(resp *http.Response, raw []byte) *HTTPError {
	return &HTTPError{Status: resp.StatusCode, RetryAfter: retryAfter(resp.Header.Get(headerRetryAfter)), Body: string(raw)}
}

// retryAfter reads a Retry-After value, in seconds or as an HTTP date.
func retryAfter(v string) time.Duration {
	v = strings.TrimSpace(v)
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		return max(0, time.Until(t))
	}
	return 0
}

// Retry runs do, retrying server errors and network failures up to retries
// more times, with exponential backoff that honors Retry-After. A rate limit
// (429) is waited out for up to rateLimitWait in all, without counting against
// retries: a hosted plan's usage window resets after hours, and failing the
// request fails the game. With no rateLimitWait a rate limit is retried like
// any other transient failure.
func Retry(ctx context.Context, retries int, rateLimitWait time.Duration, do func() (Reply, error)) (Reply, error) {
	var (
		failures, limits int
		waited           time.Duration
	)
	for {
		reply, err := do()
		if err == nil {
			return reply, nil
		}
		if !retryable(err) {
			slog.Warn("chat request failed", "error", err)
			return Reply{}, err
		}
		var delay time.Duration
		if limitWait, ok := rateLimitDelay(err, limits); ok && rateLimitWait > 0 {
			if waited >= rateLimitWait {
				return Reply{}, fmt.Errorf("chat request still rate limited after waiting %s: %w", waited.Round(time.Second), err)
			}
			limits++
			delay = min(limitWait, rateLimitWait-waited)
			waited += delay
			slog.Warn("chat request rate limited, waiting", "wait", delay.Round(time.Second), "waited", waited.Round(time.Second), "of", rateLimitWait, "error", err)
		} else {
			failures++
			slog.Warn("chat request failed", "attempt", failures, "of", retries+1, "error", err)
			if failures > retries {
				return Reply{}, fmt.Errorf("chat request failed after %d attempts: %w", failures, err)
			}
			delay = backoff(failures, err)
		}
		if !sleep(ctx, delay) {
			return Reply{}, ctx.Err()
		}
	}
}

func retryable(err error) bool {
	var he *HTTPError
	if errors.As(err, &he) {
		return he.Status == http.StatusTooManyRequests || he.Status >= 500
	}
	return true
}

// rateLimitDelay is how long to wait after the request's limits-th rate limit
// when err is one: its Retry-After, or a backoff that grows to minutes.
func rateLimitDelay(err error, limits int) (time.Duration, bool) {
	var he *HTTPError
	if !errors.As(err, &he) || he.Status != http.StatusTooManyRequests {
		return 0, false
	}
	if he.RetryAfter > 0 {
		return he.RetryAfter, true
	}
	return min(RateLimitBackoffBase*time.Duration(1<<min(limits, 16)), rateLimitBackoffMax), true
}

func backoff(retry int, cause error) time.Duration {
	var he *HTTPError
	if errors.As(cause, &he) && he.RetryAfter > 0 {
		return he.RetryAfter
	}
	return min(BackoffBase*time.Duration(1<<(retry-1)), backoffMax)
}

// sleep waits d and reports whether ctx is still live.
func sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
