package session

import (
	"fmt"
	"io"
	"strings"

	"mcbench/constants"
)

// writeLive renders one decision and the board after it to opts.Live.
func (s *Session) writeLive(d *Decision, chosen Option, reasoning string, events []string) {
	w := s.opts.Live
	if w == nil {
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\n=== decision %d · round %d · %s ===\n", len(s.Choices), s.g.S.Round, d.Kind)
	fmt.Fprintf(&b, "%s\n", d.Prompt)
	fmt.Fprintf(&b, "> chose: %s\n", chosen.Text)
	if reasoning != "" {
		fmt.Fprintf(&b, "  why: %s\n", oneLine(reasoning))
	}
	for _, e := range events {
		fmt.Fprintf(&b, "  · %s\n", e)
	}
	b.WriteString("\n")
	b.WriteString(renderView(s.View()))
	if st := s.Status(); st != AwaitingDecision {
		fmt.Fprintf(&b, "GAME OVER: %s\n", st)
	}
	io.WriteString(w, b.String())
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func renderView(v View) string {
	var b strings.Builder
	fmt.Fprintf(&b, "round %d · %s phase\n", v.Round, v.Phase)
	fmt.Fprintf(&b, "hero: %s (%s) %d/%d HP", v.Hero.Name, v.Hero.Form, v.Hero.HP, v.Hero.MaxHP)
	if v.Hero.Form == constants.TypeHero {
		fmt.Fprintf(&b, " · atk %d thw %d def %d", v.Hero.ATK, v.Hero.THW, v.Hero.DEF)
	} else {
		fmt.Fprintf(&b, " · rec %d", v.Hero.REC)
	}
	if v.Hero.Exhausted {
		b.WriteString(" · exhausted")
	}
	if len(v.Hero.Statuses) > 0 {
		fmt.Fprintf(&b, " · %s", strings.Join(v.Hero.Statuses, ","))
	}
	b.WriteString("\n")
	fmt.Fprintf(&b, "hand (%d): %s\n", len(v.Hand), cardList(v.Hand))
	if s := cardList(v.InPlay); s != "" {
		fmt.Fprintf(&b, "in play: %s\n", s)
	}
	fmt.Fprintf(&b, "villain: %s %d/%d HP · atk %d sch %d%s\n", v.Villain.Name, v.Villain.HP, v.Villain.MaxHP, v.Villain.ATK, v.Villain.SCH, tags(v.Villain))
	fmt.Fprintf(&b, "main scheme: %s %d/%d%s\n", v.MainScheme.Name, v.MainScheme.Threat, v.MainScheme.Threshold, perRound(v.MainScheme))
	for _, s := range v.SideSchemes {
		fmt.Fprintf(&b, "side scheme: %s %d threat%s%s\n", s.Name, s.Threat, icons(s), perRound(s))
	}
	for _, m := range v.Minions {
		fmt.Fprintf(&b, "minion: %s %d/%d HP · atk %d sch %d%s\n", m.Name, m.HP, m.MaxHP, m.ATK, m.SCH, tags(m))
	}
	fmt.Fprintf(&b, "deck %d · discard %d · encounter deck %d · encounter discard %d · face-down %d",
		v.DeckSize, len(v.Discard), v.EncounterDeckSize, len(v.EncounterDiscard), v.FaceDownEncounters)
	if v.AccelerationTokens > 0 {
		fmt.Fprintf(&b, " · acceleration %d", v.AccelerationTokens)
	}
	b.WriteString("\n")
	return b.String()
}

func cardList(cs []Card) string {
	parts := make([]string, len(cs))
	for i, c := range cs {
		name := c.Name
		if c.Label != "" {
			name = c.Label
		}
		if c.Cost != nil {
			name = fmt.Sprintf("%s(%d)", name, *c.Cost)
		}
		if len(c.Resources) > 0 {
			name += "[" + strings.Join(c.Resources, "/") + "]"
		}
		if c.Exhausted {
			name += "*"
		}
		parts[i] = name
	}
	return strings.Join(parts, ", ")
}

func tags(e Enemy) string {
	var t []string
	t = append(t, e.Keywords...)
	t = append(t, e.Statuses...)
	if len(t) == 0 {
		return ""
	}
	return " · " + strings.Join(t, ",")
}

func icons(s Scheme) string {
	if len(s.Icons) == 0 {
		return ""
	}
	return " · " + strings.Join(s.Icons, ",")
}

func perRound(s Scheme) string {
	if s.PerRound == 0 {
		return ""
	}
	return fmt.Sprintf(" (+%d/round)", s.PerRound)
}
