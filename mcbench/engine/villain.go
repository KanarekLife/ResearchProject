package engine

import "fmt"

// villainPhase runs the villain phase of the round (Rules Reference order).
func (g *Game) villainPhase() {
	s := g.S
	s.Phase = PhaseVillain
	g.Logf("--- Villain phase, round %d ---", s.Round)
	g.phaseBegins(PhaseVillain)
	if g.halted {
		return
	}

	// 1. Place threat on the main scheme.
	n := s.MainScheme.Face().Escalation + s.AccelTokens
	for _, ss := range s.SideSchemes {
		n += ss.Face().Acceleration
	}

	// 2-3. Villain activates, then each engaged minion.
	steps := []func(){
		func() { g.PlaceThreatWindowed(s.MainScheme, n) },
		func() { g.activate(s.Villain, true) },
	}
	for _, m := range append([]*Card(nil), s.Minions...) {
		minion := m
		steps = append(steps, func() {
			if contains(s.Minions, minion) {
				g.activate(minion, false)
			}
		})
	}
	// 4-5. Deal and reveal encounter cards.
	steps = append(steps, g.dealAndRevealEncounters)
	g.Do(steps...)
}

func (g *Game) endRound() {
	s := g.S
	g.fireForced(&Event{Trigger: TrigRoundEnd})
	s.Round++
	s.OncePerRound = map[string]bool{}
	g.Logf("--- End of round. Round %d begins ---", s.Round)
	s.Phase = PhasePlayer
	g.Do(g.playerTurn)
}

// activate makes an enemy attack (hero form) or scheme (alter-ego form).
func (g *Game) activate(enemy *Card, isVillain bool) {
	if g.IsHero() {
		g.EnemyAttack(enemy, isVillain, nil)
	} else {
		g.EnemyScheme(enemy, isVillain)
	}
}

// drawEncounter takes the top encounter card, reshuffling the discard pile
// into the deck (with an acceleration token) when the deck is empty.
func (g *Game) drawEncounter() *Card {
	s := g.S
	if len(s.EncDeck) == 0 {
		if len(s.EncDiscard) == 0 {
			return nil
		}
		s.EncDeck, s.EncDiscard = s.EncDiscard, nil
		g.shuffle(s.EncDeck)
		s.AccelTokens++
		g.Logf("Encounter deck reshuffled; an acceleration token is added (%d total).", s.AccelTokens)
	}
	c := s.EncDeck[0]
	s.EncDeck = s.EncDeck[1:]
	return c
}

// boost deals and immediately resolves a boost card, returning its icons.
func (g *Game) boost() int {
	c := g.drawEncounter()
	if c == nil {
		return 0
	}
	g.Logf("Boost card revealed: %s (+%d).", c.Def.Name, c.Def.Boost)
	g.S.EncDiscard = append(g.S.EncDiscard, c)
	return c.Def.Boost
}

// EnemyATK is an enemy's ATK including attachments.
func (g *Game) EnemyATK(enemy *Card) int { return enemy.Face().ATK + g.attachedATK(enemy) }

// EnemySCH is an enemy's SCH including attachments.
func (g *Game) EnemySCH(enemy *Card) int { return enemy.Face().SCH + g.attachedSCH(enemy) }

func (g *Game) attachedATK(enemy *Card) (n int) {
	for _, a := range enemy.Attached {
		n += a.Face().AttachATK
	}
	return
}

func (g *Game) attachedSCH(enemy *Card) (n int) {
	for _, a := range enemy.Attached {
		n += a.Face().AttachSCH
	}
	return
}

// AttackState is shared across the steps of one enemy attack.
type AttackState struct {
	Enemy    *Card
	Overkill bool
	boostN   int
	defender *Card // nil = undefended; hero or an ally
}

// EnemyAttack resolves an enemy attack against the hero. onDamaged is
// called with each character damaged by the attack.
func (g *Game) EnemyAttack(enemy *Card, isVillain bool, onDamaged func(*Card)) {
	s := g.S
	if enemy.Stunned {
		enemy.Stunned = false
		g.Logf("%s is stunned: its attack is replaced by discarding the stun.", enemy.Name())
		return
	}
	ev := &Event{Trigger: TrigEnemyWouldAttack, Source: enemy}
	g.fireForced(ev)
	if ev.Cancelled {
		return
	}
	atk := &AttackState{Enemy: enemy}
	if isVillain {
		g.fireForced(&Event{Trigger: TrigVillainAttacks, Source: enemy})
	}
	g.Logf("%s attacks %s.", enemy.Name(), s.Hero.Name())

	var opts []Option
	h := s.Hero
	if !h.Exhausted {
		opts = append(opts, Option{
			Key:  "defend:" + h.Name(),
			Text: fmt.Sprintf("Defend with %s (exhaust; DEF %d reduces the damage)", h.Name(), g.HeroDEF()),
			do:   func() { h.Exhausted = true; atk.defender = h },
		})
	}
	for _, a := range s.Play {
		if a.Def.Type != TypeAlly || a.Exhausted {
			continue
		}
		ally := a
		opts = append(opts, Option{
			Key:  "defend:" + g.Label(ally),
			Text: fmt.Sprintf("Defend with ally %s (exhaust; it takes all the damage, %d HP left)", g.Label(ally), ally.RemainingHP()),
			do:   func() { ally.Exhausted = true; atk.defender = ally },
		})
	}
	resolve := func() { g.resolveAttack(atk, isVillain, onDamaged) }
	if len(opts) == 0 {
		g.Do(resolve)
		return
	}
	opts = append(opts, Option{Key: "defend:none", Text: "Do not defend (take the damage)"})
	base := enemy.Face().ATK + g.attachedATK(enemy)
	prompt := fmt.Sprintf("%s attacks you (base ATK %d", enemy.Name(), base)
	if isVillain {
		prompt += ", plus a face-down boost card"
	}
	prompt += "). Declare a defender."
	g.Ask(&Decision{Kind: "defend", Prompt: prompt, Options: opts})
	g.Do(resolve)
}

func (g *Game) resolveAttack(atk *AttackState, isVillain bool, onDamaged func(*Card)) {
	s := g.S
	enemy := atk.Enemy
	if isVillain {
		atk.boostN = g.boost()
	}
	for _, a := range enemy.Attached {
		if a.Face().AttachOverkill {
			atk.Overkill = true
		}
	}
	amount := enemy.Face().ATK + g.attachedATK(enemy) + atk.boostN
	g.Logf("%s's attack has %d ATK.", enemy.Name(), amount)

	finish := func() {
		g.fireForced(&Event{Trigger: TrigAttackEnded, Source: enemy})
	}
	toHero := 0
	switch d := atk.defender; {
	case d == nil:
		toHero = amount
	case d == s.Hero:
		toHero = max(0, amount-g.HeroDEF())
		g.Logf("%s defends, reducing the damage by %d.", s.Hero.Name(), g.HeroDEF())
	default:
		dealt, excess := g.DamageAlly(d, amount)
		if dealt > 0 && onDamaged != nil && contains(s.Play, d) {
			onDamaged(d)
		}
		if atk.Overkill && excess > 0 {
			g.Logf("Overkill: %d excess damage goes to %s.", excess, s.Hero.Name())
			toHero = excess
		}
	}
	if toHero <= 0 {
		g.Do(finish)
		return
	}
	ev := &Event{Trigger: TrigWouldTakeAttackDamage, Source: enemy, Target: s.Hero, Amount: toHero}
	g.playerWindow(ev, fmt.Sprintf("You are about to take %d damage from %s's attack.", toHero, enemy.Name()), func() {
		if g.DamageHero(ev.Amount) > 0 && onDamaged != nil {
			onDamaged(s.Hero)
		}
		g.Do(finish)
	})
}

func (g *Game) EnemyScheme(enemy *Card, isVillain bool) {
	if enemy.Confused {
		enemy.Confused = false
		g.Logf("%s is confused: its scheme is replaced by discarding the confuse.", enemy.Name())
		return
	}
	n := enemy.Face().SCH + g.attachedSCH(enemy)
	if isVillain {
		n += g.boost()
	}
	g.Logf("%s schemes for %d.", enemy.Name(), n)
	if !isVillain {
		g.PlaceThreatWindowed(g.S.MainScheme, n)
		return
	}
	ev := &Event{Trigger: TrigVillainSchemes, Source: enemy, Amount: n}
	g.playerWindow(ev, fmt.Sprintf("%s is scheming for %d threat.", enemy.Name(), n), func() {
		g.PlaceThreatWindowed(g.S.MainScheme, max(0, ev.Amount))
	})
}

// PlaceThreatWindowed places threat during the villain phase, first
// offering interrupts such as Great Responsibility.
func (g *Game) PlaceThreatWindowed(scheme *Card, n int) {
	if n <= 0 {
		return
	}
	ev := &Event{Trigger: TrigThreatWouldBePlaced, Target: scheme, Amount: n}
	g.playerWindow(ev, fmt.Sprintf("%d threat is about to be placed on %s (now %d/%d).", n, scheme.Name(), scheme.Threat, scheme.Face().TargetThreat), func() {
		g.PlaceThreat(scheme, ev.Amount)
	})
}

// dealAndRevealEncounters deals 1 + hazard encounter cards and reveals them,
// together with any cards dealt earlier in the round.
func (g *Game) dealAndRevealEncounters() {
	s := g.S
	n := 1
	for _, ss := range s.SideSchemes {
		n += ss.Face().Hazard
	}
	dealt := s.Dealt
	s.Dealt = nil
	for i := 0; i < n; i++ {
		if c := g.drawEncounter(); c != nil {
			dealt = append(dealt, c)
		}
	}
	var steps []func()
	for _, c := range dealt {
		card := c
		steps = append(steps, func() { g.Reveal(card) })
	}
	g.Do(steps...)
}

// Surge makes the current reveal reveal one more encounter card afterwards.
func (g *Game) Surge() {
	g.Logf("Surge!")
	g.Do(func() {
		if c := g.drawEncounter(); c != nil {
			g.Reveal(c)
		}
	})
}

// Reveal resolves an encounter card.
func (g *Game) Reveal(c *Card) {
	s := g.S
	d := c.Def
	g.Logf("Encounter card revealed: %s (%s).", d.Name, d.Type)
	onReveal := func() {
		if d.Script != nil && d.Script.OnReveal != nil {
			d.Script.OnReveal(g, c)
		}
	}
	var steps []func()
	switch d.Type {
	case TypeTreachery:
		ev := &Event{Trigger: TrigTreacheryRevealed, Source: c}
		steps = append(steps,
			func() {
				g.playerWindow(ev, fmt.Sprintf("Treachery %s is being revealed: %s", d.Name, d.Text), func() {})
			},
			func() {
				if ev.Cancelled {
					g.Logf("The When Revealed effect of %s is cancelled.", d.Name)
				} else {
					onReveal()
				}
			},
			func() { s.EncDiscard = append(s.EncDiscard, c) },
		)
	case TypeMinion:
		steps = append(steps, func() {
			g.EngageMinion(c)
			onReveal()
		})
	case TypeSideScheme:
		steps = append(steps, func() {
			c.Threat = d.StartingThreat
			s.SideSchemes = append(s.SideSchemes, c)
			g.Logf("Side scheme %s enters play with %d threat.", d.Name, c.Threat)
			onReveal()
		})
	case TypeAttachment:
		steps = append(steps, func() {
			s.Villain.Attached = append(s.Villain.Attached, c)
			g.Logf("%s attaches to %s.", d.Name, s.Villain.Name())
			onReveal()
		})
	case TypeObligation:
		// The card's script decides where it goes (removed or discarded).
		steps = append(steps, onReveal)
	default:
		steps = append(steps, onReveal, func() { s.EncDiscard = append(s.EncDiscard, c) })
	}
	if d.Surge {
		steps = append(steps, g.Surge)
	}
	g.Do(steps...)
}
