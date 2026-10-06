// Command mcbench runs the Marvel Champions full-game benchmark.
package main

import (
	"fmt"
	"os"
)

const usage = `mcbench - Marvel Champions full-game benchmark

Commands:
  list      list scenarios
  validate  play every scenario seed with the scripted players (engine smoke test)
  play      play a scenario yourself in the terminal
  run       play scenarios with a player and write game records
  report    summarize games.jsonl files

Run "mcbench <command> -h" for flags.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	setupLogging(defaultLogLevel) // run overrides it with -log-level
	cmds := map[string]func([]string) error{
		cmdNameList: cmdList, cmdNameValidate: cmdValidate,
		cmdNamePlay: cmdPlay, cmdNameRun: cmdRun, cmdNameReport: cmdReport,
	}
	cmd, ok := cmds[os.Args[1]]
	if !ok {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	if err := cmd(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
