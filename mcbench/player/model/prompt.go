package model

import _ "embed"

// systemPrompt is the fixed answer-format prompt. Instruction documents are
// loaded at runtime and appended; this file never contains game knowledge.
//
//go:embed system.md
var systemPrompt string

// instructionsHeader introduces the instruction documents appended to the
// system prompt.
const instructionsHeader = "\n\n# Instructions\n\n"

// system builds the system prompt: the fixed format, then the instruction
// documents (the independent variable).
func (m *Model) system() string {
	if m.Instructions.Text == "" {
		return systemPrompt
	}
	return systemPrompt + instructionsHeader + m.Instructions.Text
}
