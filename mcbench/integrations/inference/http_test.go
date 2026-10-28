package inference

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func rateLimited(retryAfter time.Duration) error {
	return &HTTPError{Status: http.StatusTooManyRequests, RetryAfter: retryAfter}
}

// A usage window resets after hours: rate limits must not spend the retries
// meant for server errors, or a game fails on the first one.
func TestRetryWaitsOutRateLimitsWithoutSpendingRetries(t *testing.T) {
	BackoffBase, RateLimitBackoffBase = time.Millisecond, time.Millisecond
	calls := 0
	_, err := Retry(context.Background(), 1, time.Minute, func() (Reply, error) {
		if calls++; calls <= 5 {
			return Reply{}, rateLimited(0)
		}
		return Reply{}, nil
	})
	if err != nil || calls != 6 {
		t.Fatalf("err=%v after %d calls, want success on call 6", err, calls)
	}
}

func TestRetryGivesUpWhenRateLimitWaitIsSpent(t *testing.T) {
	calls := 0
	_, err := Retry(context.Background(), 3, 5*time.Millisecond, func() (Reply, error) {
		calls++
		return Reply{}, rateLimited(2 * time.Millisecond)
	})
	if err == nil || calls != 4 || !strings.Contains(err.Error(), "still rate limited") {
		t.Fatalf("err=%v after %d calls, want to give up after 3 waits (2+2+1ms)", err, calls)
	}
}

func TestRetryCountsRateLimitsWithoutWait(t *testing.T) {
	BackoffBase = time.Millisecond
	calls := 0
	_, err := Retry(context.Background(), 2, 0, func() (Reply, error) {
		calls++
		return Reply{}, rateLimited(0)
	})
	if err == nil || calls != 3 {
		t.Fatalf("err=%v after %d calls, want failure after 3 attempts", err, calls)
	}
}

func TestRetryAfterReadsSecondsAndDates(t *testing.T) {
	if got := retryAfter("120"); got != 2*time.Minute {
		t.Errorf("seconds: %v", got)
	}
	got := retryAfter(time.Now().Add(time.Hour).UTC().Format(http.TimeFormat))
	if got < 59*time.Minute || got > time.Hour {
		t.Errorf("date: %v, want about an hour", got)
	}
	if got := retryAfter("soon"); got != 0 {
		t.Errorf("junk: %v", got)
	}
}
