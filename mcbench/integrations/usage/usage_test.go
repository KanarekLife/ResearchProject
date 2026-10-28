package usage

import (
	"encoding/json"
	"testing"
	"time"
)

// The replies as the providers send them; a window that is not "ok" or is
// exceeded is full whatever its percent says.
func TestWindowsReadProviderReplies(t *testing.T) {
	var oc openCodeUsage
	json.Unmarshal([]byte(`{"usage":{"rolling":{"status":"ok","percent":12,"resetsAt":"2026-10-28T19:06:24.000Z"},
		"weekly":{"status":"limited","percent":97,"resetsAt":"2026-11-02T00:00:00.000Z"}}}`), &oc)
	got := map[string]Window{}
	for _, w := range oc.windows() {
		got[w.Name] = w
	}
	if got["rolling"].Percent != 12 || got["weekly"].Percent != 100 || !got["rolling"].ResetsAt.Equal(time.Date(2026, 10, 28, 19, 6, 24, 0, time.UTC)) {
		t.Errorf("opencode windows %+v", got)
	}

	var cc commandCodeCredits
	json.Unmarshal([]byte(`{"windowLimits":{"fiveHour":{"used":7,"cap":14,"exceeded":false,"resetAt":1793491200000},
		"weekly":{"used":1,"cap":35,"exceeded":true,"resetAt":0}}}`), &cc)
	got = map[string]Window{}
	for _, w := range cc.windows() {
		got[w.Name] = w
	}
	if got["five_hour"].Percent != 50 || got["weekly"].Percent != 100 || got["five_hour"].ResetsAt.IsZero() || !got["weekly"].ResetsAt.IsZero() {
		t.Errorf("commandcode windows %+v", got)
	}
}
