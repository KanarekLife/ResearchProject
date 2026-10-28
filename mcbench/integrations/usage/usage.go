// Package usage reads how much of a hosted plan's usage windows is spent, so
// that a worker takes a job only while its provider has room.
package usage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
)

// Plans whose usage this package reads, as a provider's `usage` names them.
const (
	PlanOpenCode    = "opencode"
	PlanCommandCode = "commandcode"
)

// The usage endpoints. Command Code's is undocumented: its CLI reads it.
const (
	openCodeURL    = "https://opencode.ai/zen/go/v1/usage"
	commandCodeURL = "https://api.commandcode.ai/alpha/billing/credits"
)

const requestTimeout = 30 * time.Second

var httpClient = &http.Client{Timeout: requestTimeout}

// Window is one usage window of a plan, e.g. the rolling five hours.
type Window struct {
	Name     string
	Percent  float64   // spent; 100 once the window is exhausted
	ResetsAt time.Time // zero when the plan does not say
}

// Check reads the usage windows of the plan apiKey belongs to.
func Check(ctx context.Context, plan, apiKey string) ([]Window, error) {
	switch plan {
	case PlanOpenCode:
		var r openCodeUsage
		if err := get(ctx, openCodeURL, apiKey, &r); err != nil {
			return nil, err
		}
		return sorted(r.windows()), nil
	case PlanCommandCode:
		var r commandCodeCredits
		if err := get(ctx, commandCodeURL, apiKey, &r); err != nil {
			return nil, err
		}
		return sorted(r.windows()), nil
	}
	return nil, fmt.Errorf("unknown usage plan %q: use %s or %s", plan, PlanOpenCode, PlanCommandCode)
}

func sorted(ws []Window) []Window {
	slices.SortFunc(ws, func(a, b Window) int { return strings.Compare(a.Name, b.Name) })
	return ws
}

func get(ctx context.Context, url, apiKey string, into any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("usage API HTTP %d: %.300s", resp.StatusCode, raw)
	}
	return json.Unmarshal(raw, into)
}

// openCodeUsage is OpenCode Go's GET /usage.
type openCodeUsage struct {
	Usage map[string]struct {
		Status   string    `json:"status"`
		Percent  float64   `json:"percent"`
		ResetsAt time.Time `json:"resetsAt"`
	} `json:"usage"`
}

// openCodeOK is the status of a window that still has room.
const openCodeOK = "ok"

func (r openCodeUsage) windows() []Window {
	var ws []Window
	for name, u := range r.Usage {
		w := Window{Name: name, Percent: u.Percent, ResetsAt: u.ResetsAt}
		if u.Status != openCodeOK {
			w.Percent = 100
		}
		ws = append(ws, w)
	}
	return ws
}

// commandCodeCredits is the part of Command Code's credits we read.
type commandCodeCredits struct {
	WindowLimits struct {
		FiveHour commandCodeWindow `json:"fiveHour"`
		Weekly   commandCodeWindow `json:"weekly"`
	} `json:"windowLimits"`
}

type commandCodeWindow struct {
	Used     float64 `json:"used"`
	Cap      float64 `json:"cap"`
	Exceeded bool    `json:"exceeded"`
	ResetAt  int64   `json:"resetAt"` // Unix milliseconds; 0 before the window starts
}

func (r commandCodeCredits) windows() []Window {
	var ws []Window
	for name, u := range map[string]commandCodeWindow{"five_hour": r.WindowLimits.FiveHour, "weekly": r.WindowLimits.Weekly} {
		w := Window{Name: name}
		switch {
		case u.Exceeded:
			w.Percent = 100
		case u.Cap > 0:
			w.Percent = 100 * u.Used / u.Cap
		}
		if u.ResetAt > 0 {
			w.ResetsAt = time.UnixMilli(u.ResetAt).UTC()
		}
		ws = append(ws, w)
	}
	return ws
}
