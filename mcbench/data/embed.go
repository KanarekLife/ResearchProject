// Package data embeds the card definitions. Scenarios, decks, villains,
// encounter sets and instruction documents are loaded from disk at run time
// (see the -root and -scenarios flags).
package data

import "embed"

// FS holds the embedded card definition files.
//
//go:embed cards/*.yaml
var FS embed.FS
