package inference

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	sseEventPrefix = "event:"
	sseDataPrefix  = "data:"
	// maxEventLine bounds one line of the stream; tool arguments can be long.
	maxEventLine = 1 << 24
)

// PostSSE posts body to url and calls onEvent for each server-sent event of the
// reply: its event name (empty when the stream gives none) and its data. A
// non-200 answer is an *HTTPError, as for PostJSON. Streaming keeps the
// connection busy, so a gateway that cuts a quiet request after a couple of
// minutes lets a slow reply through. onEvent may return a stop error, which
// ends the read and is returned.
func PostSSE(ctx context.Context, url string, headers map[string]string, body []byte, onEvent func(event string, data []byte) error) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		he := &HTTPError{Status: resp.StatusCode, Body: string(raw)}
		if secs, err := strconv.Atoi(strings.TrimSpace(resp.Header.Get("Retry-After"))); err == nil {
			he.RetryAfter = time.Duration(secs) * time.Second
		}
		return he
	}
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(nil, maxEventLine)
	var event string
	var data []byte
	flush := func() error {
		defer func() { event, data = "", nil }()
		if data == nil {
			return nil
		}
		return onEvent(event, data)
	}
	for sc.Scan() {
		line := sc.Text()
		switch {
		case line == "":
			if err := flush(); err != nil {
				return err
			}
		case strings.HasPrefix(line, sseEventPrefix):
			event = strings.TrimSpace(strings.TrimPrefix(line, sseEventPrefix))
		case strings.HasPrefix(line, sseDataPrefix):
			d := strings.TrimSpace(strings.TrimPrefix(line, sseDataPrefix))
			data = append(append(data, d...), '\n')
		}
		// anything else is a comment or keep-alive
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return flush()
}
