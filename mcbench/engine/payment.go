package engine

import (
	"sort"
	"strings"
)

// paySource is one thing that can produce resources: a hand card or a
// resource ability of a card in play.
type paySource struct {
	card     *Card
	fromHand bool
	gives    []Resource
}

func (p paySource) label() string {
	if p.fromHand {
		return p.card.Name()
	}
	return p.card.Name() + "(ability)"
}

// Payment is a set of sources that together pay a cost.
type Payment struct {
	sources []paySource
}

// Label is a stable key, e.g. "Energy+Genius". Empty for a free payment.
func (p Payment) Label() string {
	names := make([]string, len(p.sources))
	for i, s := range p.sources {
		names[i] = s.label()
	}
	sort.Strings(names)
	return strings.Join(names, "+")
}

// Describe is the human-readable suffix for option text.
func (p Payment) Describe() string {
	if len(p.sources) == 0 {
		return ""
	}
	parts := make([]string, len(p.sources))
	for i, s := range p.sources {
		res := make([]string, len(s.gives))
		for j, r := range s.gives {
			res[j] = string(r)
		}
		where := "discard"
		if !s.fromHand {
			where = "use"
		}
		parts[i] = where + " " + s.card.Name() + " [" + strings.Join(res, ",") + "]"
	}
	return ", paying: " + strings.Join(parts, "; ")
}

// Spend discards the hand cards and uses the abilities.
func (p Payment) Spend(g *Game) {
	g.LastPaid = nil
	for _, s := range p.sources {
		g.LastPaid = append(g.LastPaid, s.gives...)
	}
	for _, s := range p.sources {
		if s.fromHand {
			g.moveFromHand(s.card)
			g.S.Discard = append(g.S.Discard, s.card)
		} else {
			s.card.Face().Script.ResourceAbility.Use(g, s.card)
		}
	}
}

// paySources lists what can pay for `paying` (nil for an ability cost).
func (g *Game) paySources(exclude *Card, paying *CardDef) []paySource {
	var out []paySource
	for _, c := range g.S.Hand {
		if c == exclude || len(c.Def.Resources) == 0 {
			continue
		}
		gives := c.Def.Resources
		if c.Def.DoubleFor != "" && paying != nil && paying.Aspect == c.Def.DoubleFor {
			gives = append(append([]Resource(nil), gives...), gives...)
		}
		out = append(out, paySource{card: c, fromHand: true, gives: gives})
	}
	g.inPlay(func(c *Card) {
		sc := c.Face().Script
		if sc == nil || sc.ResourceAbility == nil {
			return
		}
		if sc.ResourceAbility.Usable == nil || sc.ResourceAbility.Usable(g, c) {
			out = append(out, paySource{card: c, gives: []Resource{sc.ResourceAbility.Gives}})
		}
	})
	return out
}

// covers reports whether the pooled resources pay `cost` including the
// specific `need` symbols (each need symbol counts toward cost).
func covers(pool []Resource, cost int, need []Resource) bool {
	total := cost
	if len(need) > total {
		total = len(need)
	}
	if len(pool) < total {
		return false
	}
	avail := map[Resource]int{}
	for _, r := range pool {
		avail[r]++
	}
	for _, r := range need {
		if avail[r] > 0 {
			avail[r]--
		} else if avail[Wild] > 0 {
			avail[Wild]--
		} else {
			return false
		}
	}
	return true
}

// paymentsFor lists every minimal way to pay `cost` for `card` (excluded
// from the pool; paying is the card being played, nil for an ability).
// Overpaying is never offered: removing any source from a listed payment
// would make it insufficient. Equivalent payments are deduplicated.
func (g *Game) paymentsFor(card *Card, paying *CardDef, cost int, need []Resource) []Payment {
	if cost <= 0 && len(need) == 0 {
		return []Payment{{}}
	}
	srcs := g.paySources(card, paying)
	if len(srcs) > 16 {
		srcs = srcs[:16] // keeps enumeration bounded; hands are never this big
	}
	pool := func(mask int) []Resource {
		var p []Resource
		for i, s := range srcs {
			if mask&(1<<i) != 0 {
				p = append(p, s.gives...)
			}
		}
		return p
	}
	var out []Payment
	seen := map[string]bool{}
	for mask := 1; mask < 1<<len(srcs); mask++ {
		if !covers(pool(mask), cost, need) {
			continue
		}
		minimal := true
		for i := range srcs {
			if mask&(1<<i) != 0 && covers(pool(mask&^(1<<i)), cost, need) {
				minimal = false
				break
			}
		}
		if !minimal {
			continue
		}
		var p Payment
		for i, s := range srcs {
			if mask&(1<<i) != 0 {
				p.sources = append(p.sources, s)
			}
		}
		if l := p.Label(); !seen[l] {
			seen[l] = true
			out = append(out, p)
		}
	}
	return out
}

func (g *Game) moveFromHand(c *Card) {
	g.S.Hand = remove(g.S.Hand, c)
}
