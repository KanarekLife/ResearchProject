package trace

import (
	"fmt"
	"strings"

	"mcbench/constants"
	"mcbench/ui"
)

// Write renders an entry as a titled panel.
func Write(b *strings.Builder, e Entry, t ui.Theme) {
	var panel []string
	switch e.Type {
	case constants.TraceStart:
		panel = startPanel(e, t)
	case constants.TraceMessage:
		panel = messagePanel(e, t)
	case constants.TraceChoice:
		panel = choicePanel(e, t)
	case constants.TraceEnd:
		panel = endPanel(e, t)
	}
	if len(panel) > 0 {
		b.WriteString(strings.Join(panel, "\n") + "\n")
	}
}

func startPanel(e Entry, t ui.Theme) []string {
	lines := []string{fmt.Sprintf("%s · seed %d · sample %d · player %s", e.Scenario, e.Seed, e.Sample, e.Player)}
	if names := strings.Join(e.Instructions.Names, ","); names != "" {
		lines = append(lines, t.Dim(fmt.Sprintf("instructions %s@%s", names, short(e.Instructions.SHA256, 12))))
	}
	return t.Box("Game", lines, ui.Width)
}

// messagePanel renders one model reply: the tools it called, its reasoning and
// what it cost. User nudges and tool results carry no token counts, so they are
// skipped and the transcript stays about decisions.
func messagePanel(e Entry, t ui.Theme) []string {
	if e.PromptTokens == 0 && e.CompletionTokens == 0 {
		return nil
	}
	cost := fmt.Sprintf("prompt=%d completion=%d", e.PromptTokens, e.CompletionTokens)
	if e.CachedTokens > 0 {
		cost += fmt.Sprintf(" cached=%d", e.CachedTokens)
	}
	lines := []string{t.Dim(cost)}
	if e.Truncated {
		lines = append(lines, t.Red("TRUNCATED"))
	}
	if r := e.Message.Reasoning; r != "" {
		lines = append(lines, "", t.Dim(oneLine(r)))
	}
	return t.Box("Model "+describeCalls(e.Message.ToolCalls), lines, ui.Width)
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

func choicePanel(e Entry, t ui.Theme) []string {
	lines := []string{t.Green("> " + e.Text)}
	if e.Choice.Reasoning != "" {
		lines = append(lines, t.Dim("why: "+oneLine(e.Choice.Reasoning)))
	}
	for _, ev := range e.Events {
		lines = append(lines, "· "+ev)
	}
	return t.Box(fmt.Sprintf("Decision %d · round %d · %s", e.Decision, e.Choice.Round, e.Choice.Kind), lines, ui.Width)
}

func endPanel(e Entry, t ui.Theme) []string {
	status := e.Status
	if e.Error != "" {
		status += " (" + e.Error + ")"
	}
	lines := []string{t.Bold(status), fmt.Sprintf("%d rounds · %d requests · prompt=%d completion=%d",
		e.Rounds, e.Usage.Requests, e.Usage.InputTokens, e.Usage.OutputTokens)}
	if e.Score > 0 {
		lines = append(lines, fmt.Sprintf("score %.3f", e.Score))
	}
	return t.Box("Game over", lines, ui.Width)
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

func short(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
