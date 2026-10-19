// Command mcbench runs the Marvel Champions full-game benchmark.
package main

import (
	"fmt"
	"os"
)

const usage = `mcbench - Marvel Champions full-game benchmark

Commands:
  list       list scenarios
  validate   play every scenario seed with the scripted players (engine smoke test)
  play       play a scenario yourself in the terminal
  run        play scenarios with a player and write game records
  report     summarize games.jsonl files
  analyze    how players behaved in games.jsonl files: loss causes, turn use, form, defence
  view       read a run's traces: a game transcript, a live game, or final results
  board_gui  show a game trace as a board in the browser, step by step
  site       write the board and the report and analyze tables as a static website

Run "mcbench <command> -h" for flags.
`

func main() {
	if err := loadDotEnv(".env"); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	setupLogging(defaultLogLevel) // run overrides it with -log-level
	cmds := map[string]func([]string) error{
		cmdNameList: cmdList, cmdNameValidate: cmdValidate,
		cmdNamePlay: cmdPlay, cmdNameRun: cmdRun, cmdNameReport: cmdReport, cmdNameAnalyze: cmdAnalyze,
		cmdNameView: cmdView, cmdNameBoardGUI: cmdBoardGUI, cmdNameSite: cmdSite,
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
