// Package data embeds the game content: cards, scenarios, decks, villains and
// encounter sets, so a build carries exactly the game it scores. Instruction
// documents, the variable under study, stay on disk (-instructions-dir).
package data

import "embed"

// FS holds the embedded game content.
//
//go:embed cards/*.yaml scenarios/*.yaml decks/*.yaml villains/*.yaml encounter-sets/*.yaml
var FS embed.FS
