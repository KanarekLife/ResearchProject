// Package human plays a game from a terminal: it prints the state and the
// options and reads option ids.
package human

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"mcbench/game/session"
	"mcbench/player"
)

// Human plays from a terminal.
type Human struct {
	In  *bufio.Reader
	Out io.Writer
}

func (*Human) Name() string { return "human" }

func (h *Human) Play(ctx context.Context, s *session.Session) (player.Usage, error) {
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
				return player.Usage{}, err
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
	return player.Usage{}, nil
}
