package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"mcbench/game"
)

// Human plays from a terminal: it prints the state and the options and reads
// option ids.
type Human struct {
	In  *bufio.Reader
	Out io.Writer
}

func (h *Human) Name() string { return "human" }

func (h *Human) Play(ctx context.Context, s *game.Session) (Usage, error) {
	for d := s.Decision(); d != nil; d = s.Decision() {
		state, _ := json.MarshalIndent(s.View(), "", "  ")
		fmt.Fprintf(h.Out, "%s\n\n%s\n", state, d.Prompt)
		for _, o := range d.Options {
			fmt.Fprintf(h.Out, "  [%d] %s\n", o.ID, o.Text)
		}
		for {
			fmt.Fprint(h.Out, "> option: ")
			line, err := h.In.ReadString('\n')
			if err != nil {
				return Usage{}, err
			}
			id, _ := strconv.Atoi(strings.TrimSpace(line))
			events, err := s.Choose(id, "")
			if err != nil {
				fmt.Fprintln(h.Out, err)
				continue
			}
			fmt.Fprintln(h.Out, "\n"+strings.Join(events, "\n"))
			break
		}
	}
	return Usage{}, nil
}
