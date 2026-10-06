package trace

import (
	"fmt"
	"strings"
	"unicode"
)

// Write renders entries as a readable transcript.
func Write(b *strings.Builder, e Entry) {
	switch e.Type {
	case "start":
		fmt.Fprintf(b, "scenario %s · seed %d · sample %d · player %s\n", e.Scenario, e.Seed, e.Sample, e.Player)
		if names := strings.Join(e.Instructions.Names, ","); names != "" {
			fmt.Fprintf(b, "instructions %s@%s\n", names, short(e.Instructions.SHA256, 12))
		}
	case "message":
		writeMessage(b, e)
	case "choice":
		writeChoice(b, e)
	case "end":
		writeEnd(b, e)
	}
}

// writeMessage renders one model reply: the tools it called, its reasoning and
// what it cost. User nudges and tool results carry no token counts, so they are
// skipped and the transcript stays about decisions.
func writeMessage(b *strings.Builder, e Entry) {
	if e.PromptTokens == 0 && e.CompletionTokens == 0 {
		return
	}
	fmt.Fprintf(b, "  model %s prompt=%d completion=%d",
		describeCalls(e.Message.ToolCalls), e.PromptTokens, e.CompletionTokens)
	if e.CachedTokens > 0 {
		fmt.Fprintf(b, " cached=%d", e.CachedTokens)
	}
	if e.Truncated {
		b.WriteString(" TRUNCATED")
	}
	b.WriteString("\n")
	if r := e.Message.Reasoning; r != "" {
		for _, line := range wrap(indent(r, "      "), 78) {
			fmt.Fprintf(b, "%s\n", line)
		}
	}
}

// describeCalls names the tools the model called, or says it called none.
func describeCalls(calls []ToolCall) string {
	if len(calls) == 0 {
		return "called no tool"
	}
	names := make([]string, 0, len(calls))
	for _, c := range calls {
		names = append(names, c.Function.Name)
	}
	return "called " + strings.Join(names, "+")
}

func writeChoice(b *strings.Builder, e Entry) {
	fmt.Fprintf(b, "\n=== decision %d · round %d · %s ===\n", e.Decision, e.Choice.Round, e.Choice.Kind)
	fmt.Fprintf(b, "> chose: %s\n", e.Text)
	if e.Choice.Reasoning != "" {
		fmt.Fprintf(b, "  why: %s\n", oneLine(e.Choice.Reasoning))
	}
	for _, ev := range e.Events {
		fmt.Fprintf(b, "  · %s\n", ev)
	}
	b.WriteString("\n")
}

func writeEnd(b *strings.Builder, e Entry) {
	fmt.Fprintf(b, "GAME OVER: %s", e.Status)
	if e.Error != "" {
		fmt.Fprintf(b, " (%s)", e.Error)
	}
	fmt.Fprintf(b, "\n%d rounds · %d requests · prompt=%d completion=%d\n",
		e.Rounds, e.Usage.Requests, e.Usage.InputTokens, e.Usage.OutputTokens)
	if e.Score > 0 {
		fmt.Fprintf(b, "score %.3f\n", e.Score)
	}
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

// wrap breaks text into lines of at most width, on word boundaries.
func wrap(text string, width int) []string {
	var out []string
	for _, para := range strings.Split(text, "\n") {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		line := ""
		for _, word := range strings.Fields(para) {
			switch {
			case line == "":
				line = word
			case len(line)+1+len(word) <= width:
				line += " " + word
			default:
				out = append(out, line)
				line = word
			}
		}
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func indent(s, pad string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = pad + strings.TrimLeftFunc(l, unicode.IsSpace)
	}
	return strings.Join(lines, "\n")
}

func short(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
