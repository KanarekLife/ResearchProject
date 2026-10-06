package human

import (
	"fmt"
	"strings"

	"mcbench/constants"
	"mcbench/game/session"
	"mcbench/ui"
)

const (
	colGap  = 2
	barSize = 12
)

// board lays the public game state out as panels: villain beside hero, then
// schemes, minions, the table, the hand and the piles.
func board(t ui.Theme, v session.View) []string {
	half := (ui.Width - colGap) / 2
	title := fmt.Sprintf("Round %d · %s phase", v.Round, v.Phase)
	out := t.Box(title, []string{pileLine(v)}, ui.Width)
	out = append(out, ui.Side(t.Box("Villain", villainLines(t, v), half), t.Box("Hero", heroLines(t, v), half), colGap)...)
	out = append(out, t.Box("Schemes", schemeLines(t, v), ui.Width)...)
	if len(v.Minions) > 0 {
		out = append(out, t.Box("Minions", enemyLines(t, v.Minions), ui.Width)...)
	}
	if len(v.InPlay) > 0 {
		out = append(out, t.Box("In play", cardLines(t, v.InPlay), ui.Width)...)
	}
	return append(out, t.Box(fmt.Sprintf("Hand (%d)", len(v.Hand)), cardLines(t, v.Hand), ui.Width)...)
}

func pileLine(v session.View) string {
	s := fmt.Sprintf("deck %d · discard %d · encounter deck %d · encounter discard %d · face-down %d",
		v.DeckSize, len(v.Discard), v.EncounterDeckSize, len(v.EncounterDiscard), v.FaceDownEncounters)
	if v.AccelerationTokens > 0 {
		s += fmt.Sprintf(" · acceleration %d", v.AccelerationTokens)
	}
	return s
}

func hpLine(t ui.Theme, hp, maxHP int) string {
	text := fmt.Sprintf("%s %d/%d HP", ui.Bar(hp, maxHP, barSize), hp, maxHP)
	switch {
	case hp*3 <= maxHP:
		return t.Red(text)
	case hp*3 <= maxHP*2:
		return t.Yellow(text)
	}
	return t.Green(text)
}

func villainLines(t ui.Theme, v session.View) []string {
	e := v.Villain
	lines := []string{
		t.Bold(fmt.Sprintf("%s (stage %d)", e.Name, e.Stage)),
		hpLine(t, e.HP, e.MaxHP),
		fmt.Sprintf("atk %d · sch %d%s", e.ATK, e.SCH, tags(e)),
	}
	if e.Text != "" {
		lines = append(lines, t.Dim(e.Text))
	}
	for _, n := range v.NextStages {
		lines = append(lines, t.Dim(fmt.Sprintf("next: %s (%d HP)", n.Name, n.HP)))
	}
	return lines
}

func heroLines(t ui.Theme, v session.View) []string {
	h := v.Hero
	stats := fmt.Sprintf("rec %d", h.REC)
	if h.Form == constants.TypeHero {
		stats = fmt.Sprintf("atk %d · thw %d · def %d", h.ATK, h.THW, h.DEF)
	}
	state := h.Form
	if h.Exhausted {
		state += " · exhausted"
	}
	for _, s := range h.Statuses {
		state += " · " + s
	}
	lines := []string{t.Bold(h.Name), hpLine(t, h.HP, h.MaxHP), stats, state}
	if h.Ability != "" {
		lines = append(lines, t.Dim(h.Ability))
	}
	return lines
}

func schemeLines(t ui.Theme, v session.View) []string {
	m := v.MainScheme
	main := t.Bold(m.Name) + fmt.Sprintf("  threat %d/%d%s", m.Threat, m.Threshold, perRound(m))
	if m.Threshold > 0 && m.Threat*3 >= m.Threshold*2 {
		main = t.Red(main)
	}
	lines := []string{main}
	for _, s := range v.SideSchemes {
		lines = append(lines, fmt.Sprintf("side: %s  threat %d%s%s", t.Bold(s.Name), s.Threat, joined(s.Icons), perRound(s)))
	}
	return lines
}

func enemyLines(t ui.Theme, es []session.Enemy) []string {
	var lines []string
	for _, e := range es {
		lines = append(lines, fmt.Sprintf("%s  %s  atk %d · sch %d%s", t.Bold(e.Name), hpLine(t, e.HP, e.MaxHP), e.ATK, e.SCH, tags(e)))
	}
	return lines
}

func cardLines(t ui.Theme, cs []session.Card) []string {
	if len(cs) == 0 {
		return []string{t.Dim("empty")}
	}
	var lines []string
	for _, c := range cs {
		lines = append(lines, cardLine(t, c))
		if c.Text != "" {
			lines = append(lines, t.Dim("› "+c.Text))
		}
	}
	return lines
}

func cardLine(t ui.Theme, c session.Card) string {
	name := c.Name
	if c.Label != "" {
		name = c.Label
	}
	s := t.Bold(name)
	if c.Cost != nil {
		s += fmt.Sprintf(" (%d)", *c.Cost)
	}
	s += " " + t.Cyan(c.Type)
	if len(c.Resources) > 0 {
		s += " [" + strings.Join(c.Resources, "/") + "]"
	}
	if c.MaxHP > 0 {
		s += fmt.Sprintf(" · %d/%d HP · atk %d · thw %d", c.HP, c.MaxHP, c.ATK, c.THW)
	}
	if c.Counters > 0 {
		s += fmt.Sprintf(" · %d counters", c.Counters)
	}
	if c.Exhausted {
		s += " · exhausted"
	}
	for _, st := range c.Statuses {
		s += " · " + st
	}
	return s
}

func decisionLines(t ui.Theme, d *session.Decision) []string {
	lines := []string{d.Prompt, ""}
	for _, o := range d.Options {
		lines = append(lines, fmt.Sprintf("%s %s", t.Cyan(fmt.Sprintf("[%d]", o.ID)), o.Text))
	}
	return lines
}

func tags(e session.Enemy) string {
	return joined(append(append([]string{}, e.Keywords...), e.Statuses...))
}

func joined(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	return " · " + strings.Join(ss, ", ")
}

func perRound(s session.Scheme) string {
	if s.PerRound == 0 {
		return ""
	}
	return fmt.Sprintf(" (+%d/round)", s.PerRound)
}
