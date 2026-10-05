package main

import (
	"encoding/json"
	"fmt"

	"mcbench/game/session"
)

func cmdTools([]string) error {
	out, err := json.MarshalIndent(session.Tools, "", "  ")
	fmt.Println(string(out))
	return err
}
