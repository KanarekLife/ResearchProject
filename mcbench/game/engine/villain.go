package engine

import (
	"fmt"

	"mcbench/constants"
)

// villainPhase runs the villain phase of the round (Rules Reference order).
func (g *Game) villainPhase() {
	s := g.S
	g.endPhase()
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
		func() {
			stage := s.MainScheme.Def
			g.PlaceThreatThen(s.MainScheme, n, func(placed int) {
				// A stage that just advanced has not had threat placed on it.
				if placed > 0 && s.MainScheme.Def == stage {
					g.fireForced(&Event{Trigger: TrigStepOneThreat, Target: s.MainScheme, Amount: placed})
				}
			})
		},
		func() { g.activate(s.Villain, true) },
	}
	// The minions engaged when step 3 begins activate, including any that
	// entered play during the villain's activation.
	steps = append(steps, func() {
		var acts []func()
		for _, m := range append([]*Card(nil), s.Minions...) {
			acts = append(acts, func() {
				if contains(s.Minions, m) {
					g.activate(m, false)
				}
			})
		}
		g.Do(acts...)
	})
	// 4-5. Deal and reveal encounter cards.
	steps = append(steps, g.dealAndRevealEncounters)
	g.Do(steps...)
}

func (g *Game) endRound() {
	s := g.S
	g.endPhase()
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
		g.EnemyScheme(enemy, isVillain, nil)
	}
}

// EnemyATK is an enemy's ATK including attachments and other modifiers.
func (g *Game) EnemyATK(enemy *Card) int {
	n := enemy.Face().ATK + g.attachedATK(enemy) + g.statBonus(enemy, constants.StatATK)
	if enemy.Drone != nil {
		n += g.droneBonus(func(d *CardDef) int { return d.DroneATK })
	}
	return n
}

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

// EnemyScheme makes an enemy scheme. after, if set, is called with the
// threat placed.
func (g *Game) EnemyScheme(enemy *Card, isVillain bool, after func(placed int)) {
	if enemy.Confused {
		enemy.Confused = false
		g.Logf("%s is confused: its scheme is replaced by discarding the confuse.", enemy.Name())
		return
	}
	n := g.EnemySCH(enemy)
	if !isVillain {
		g.Logf("%s schemes for %d.", enemy.Name(), n)
		g.PlaceThreatThen(g.S.MainScheme, n, after)
		return
	}
	boost := g.drawEncounter()
	n += g.flipBoost(boost)
	g.Do(func() { g.boostAbility(boost, &Event{Source: enemy}) }, func() {
		g.Logf("%s schemes for %d.", enemy.Name(), n)
		ev := &Event{Trigger: TrigVillainSchemes, Source: enemy, Amount: n}
		g.playerWindow(ev, fmt.Sprintf("%s is scheming for %d threat.", enemy.Name(), n), func() {
			g.PlaceThreatThen(g.S.MainScheme, max(0, ev.Amount), func(placed int) {
				g.fireForced(&Event{Trigger: TrigVillainSchemed, Source: enemy, Amount: placed})
				if after != nil {
					after(placed)
				}
			})
		})
	})
}

// PlaceThreatWindowed places threat, first offering interrupts such as Great
// Responsibility. The window is scheduled as a step, so effects that follow
// it in the same list resolve after the player answers.
func (g *Game) PlaceThreatWindowed(scheme *Card, n int) { g.PlaceThreatThen(scheme, n, nil) }

// PlaceThreatThen is PlaceThreatWindowed that then calls after (if set) with
// the threat actually placed.
func (g *Game) PlaceThreatThen(scheme *Card, n int, after func(placed int)) {
	if n <= 0 {
		return
	}
	ev := &Event{Trigger: TrigThreatWouldBePlaced, Target: scheme, Amount: n}
	g.Do(func() {
		g.playerWindow(ev, fmt.Sprintf("%d threat is about to be placed on %s (now %d/%d).", n, scheme.Name(), scheme.Threat, scheme.Face().TargetThreat), func() {
			placed := max(0, ev.Amount)
			g.PlaceThreat(scheme, placed)
			if after != nil {
				after(placed)
			}
		})
	})
}
