package engine

import (
	"fmt"

	"mcbench/constants"
)

// Label names a card uniquely for keys and option text: "Name", or
// "Name#ID" when another card in play shares the name.
func (g *Game) Label(c *Card) string {
	if g.nameCount(c.Name()) > 1 {
		return fmt.Sprintf("%s%s%d", c.Name(), constants.SepLabel, c.ID)
	}
	return c.Name()
}

func (g *Game) nameCount(name string) int {
	n := 0
	g.inPlay(func(x *Card) {
		if x.Name() == name {
			n++
		}
	})
	return n
}
