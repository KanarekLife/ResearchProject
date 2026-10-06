package engine

import (
	"fmt"

	"mcbench/constants"
)

// TurnOptions lists the legal actions in the player's main window.
func (g *Game) TurnOptions() []Option {
	var opts []Option
	for _, c := range g.S.Hand {
		opts = append(opts, g.playOptions(c)...)
	}
	opts = append(opts, g.basicOptions()...)
	opts = append(opts, g.allyOptions()...)
	opts = append(opts, g.activatedOptions()...)
	if o, ok := g.formOption(); ok {
		opts = append(opts, o)
	}
	opts = append(opts, Option{Key: constants.KeyEndTurn, Text: "End your turn", do: func() { g.Do(g.endPlayerPhase) }})
	return opts
}

func (g *Game) basicOptions() []Option {
	h := g.S.Hero
	if h.Exhausted {
		return nil
	}
	var opts []Option
	if g.IsHero() {
		for _, target := range g.Enemies() {
			opts = append(opts, g.attackOption(h, target))
		}
		for _, scheme := range g.thwartTargets(h) {
			opts = append(opts, g.thwartOption(h, scheme))
		}
		return opts
	}
	if h.Damage > 0 {
		opts = append(opts, Option{
			Key:  constants.PrefixRecover + h.Name(),
			Text: fmt.Sprintf("Recover: %s heals %d damage (exhaust %s)", h.Name(), h.Face().REC, h.Name()),
			do: g.then(func() {
				h.Exhausted = true
				g.HealHero(h.Face().REC)
			}),
		})
	}
	return opts
}

func (g *Game) allyOptions() []Option {
	var opts []Option
	for _, ally := range g.S.Play {
		if ally.Def.Type != TypeAlly || ally.Exhausted {
			continue
		}
		opts = append(opts, g.allyAttackOptions(ally)...)
		opts = append(opts, g.allyThwartOptions(ally)...)
	}
	return opts
}

// strike describes an exhausting attack or thwart by actor against target.
// Its option key is prefix + name + target.
type strike struct {
	prefix string
	actor  *Card
	name   string // how the actor appears in the option key
	target *Card
	text   string
	run    func() // the attack or thwart itself, after the actor exhausts
}

func (g *Game) strikeOption(k strike) Option {
	return Option{
		Key:  k.prefix + k.name + constants.SepTarget + g.Label(k.target),
		Text: k.text,
		do: g.then(func() {
			k.actor.Exhausted = true
			k.run()
		}),
	}
}

func (g *Game) attackOption(h *Card, target *Card) Option {
	note := ""
	if h.Stunned {
		note = " (you are stunned: this only removes the stun)"
	}
	return g.strikeOption(strike{constants.PrefixAttack, h, h.Name(), target,
		fmt.Sprintf("Basic attack: %s deals %d damage to %s (exhaust %s)%s", h.Name(), g.HeroATK(), g.Label(target), h.Name(), note),
		func() { g.Attack(h, target, g.HeroATK(), 0) }})
}

func (g *Game) thwartOption(h *Card, scheme *Card) Option {
	note := ""
	if h.Confused {
		note = " (you are confused: this only removes the confuse)"
	}
	return g.strikeOption(strike{constants.PrefixThwart, h, h.Name(), scheme,
		fmt.Sprintf("Basic thwart: %s removes %d threat from %s (exhaust %s)%s", h.Name(), g.HeroTHW(), g.Label(scheme), h.Name(), note),
		func() { g.Thwart(h, scheme, g.HeroTHW(), 0) }})
}

// thwartTargets lists the schemes actor may make a basic thwart against. A
// confused character may attempt one (only removing the confuse) even when
// no scheme has threat.
func (g *Game) thwartTargets(actor *Card) []*Card {
	if schemes := g.ThwartableSchemes(); len(schemes) > 0 || !actor.Confused {
		return schemes
	}
	return append([]*Card{g.S.MainScheme}, g.S.SideSchemes...)
}

func (g *Game) allyAttackOptions(ally *Card) []Option {
	var opts []Option
	for _, target := range g.Enemies() {
		opts = append(opts, g.strikeOption(strike{constants.PrefixAttack, ally, g.Label(ally), target,
			fmt.Sprintf("Ally attack: %s deals %d damage to %s (exhaust; takes %d consequential damage)", g.Label(ally), g.AllyATK(ally), g.Label(target), ally.Def.AtkCons),
			func() { g.Attack(ally, target, g.AllyATK(ally), ally.Def.AtkCons) }}))
	}
	return opts
}

func (g *Game) allyThwartOptions(ally *Card) []Option {
	var opts []Option
	for _, scheme := range g.thwartTargets(ally) {
		opts = append(opts, g.strikeOption(strike{constants.PrefixThwart, ally, g.Label(ally), scheme,
			fmt.Sprintf("Ally thwart: %s removes %d threat from %s (exhaust; takes %d consequential damage)", g.Label(ally), g.AllyTHW(ally), g.Label(scheme), ally.Def.ThwCons),
			func() { g.Thwart(ally, scheme, g.AllyTHW(ally), ally.Def.ThwCons) }}))
	}
	return opts
}

func (g *Game) activatedOptions() []Option {
	var opts []Option
	g.inPlay(func(c *Card) {
		sc := c.Face().Script
		if sc == nil || sc.Actions == nil {
			return
		}
		for _, o := range sc.Actions(g, c) {
			o.do = g.then(o.do)
			opts = append(opts, o)
		}
	})
	return opts
}

func (g *Game) formOption() (Option, bool) {
	h := g.S.Hero
	if g.S.FormChanged {
		return Option{}, false
	}
	other := h.Def.Name
	if !h.Flipped {
		other = h.Def.Back.Name
	}
	return Option{
		Key:  constants.PrefixChangeForm + other,
		Text: fmt.Sprintf("Change form to %s (once per turn)", other),
		do: g.then(func() {
			from := h.Name()
			h.Flipped = !h.Flipped
			g.S.FormChanged = true
			g.Logf("%s changes form to %s.", from, h.Name())
		}),
	}, true
}

// NewAction builds an Option for a card's Actions hook. The turn continues
// automatically after it resolves.
func NewAction(key, text string, do func()) Option {
	return Option{Key: key, Text: text, do: do}
}

func (g *Game) NewActionPaid(src *Card, key, text string, cost int, need []Resource, do func()) []Option {
	var out []Option
	for _, p := range g.paymentsFor(src, nil, cost, need) {
		pay := p
		out = append(out, Option{Key: key, Pay: pay.Label(), Text: text + pay.Describe(), do: func() {
			pay.Spend(g)
			do()
		}})
	}
	return out
}

func (g *Game) playOptions(c *Card) []Option {
	d := c.Def
	sc := d.Script
	switch d.Type {
	case TypeResource:
		return nil
	case TypeEvent:
		if sc == nil || sc.OnPlay == nil || sc.PlayWindow != "" {
			return nil
		}
	}
	if d.Unique && g.uniqueInPlay(d.Name) {
		return nil
	}
	if d.MaxInPlay > 0 && g.countInPlay(d.Name) >= d.MaxInPlay {
		return nil
	}
	if sc != nil && sc.CanPlay != nil && !sc.CanPlay(g, c) {
		return nil
	}
	targets := []*Card{nil}
	if sc != nil && sc.Targets != nil {
		targets = sc.Targets(g, c)
		if len(targets) == 0 {
			return nil
		}
	}
	var opts []Option
	for _, t := range targets {
		opts = append(opts, g.payToPlay(c, d, t)...)
	}
	return opts
}

func (g *Game) payToPlay(c *Card, d *CardDef, target *Card) []Option {
	key := constants.PrefixPlay + d.Name
	text := "Play " + d.Name
	if target != nil {
		key += constants.SepTarget + g.Label(target)
		text += " targeting " + g.Label(target)
	}
	var opts []Option
	for _, p := range g.paymentsFor(c, d, max(0, d.Cost-g.S.CostReduction), nil) {
		pay := p
		opts = append(opts, Option{
			Key:  key,
			Pay:  pay.Label(),
			Text: text + pay.Describe(),
			do: g.then(func() {
				pay.Spend(g)
				g.playCard(c, target)
			}),
		})
	}
	return opts
}
