package engine

import (
	"fmt"
)

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

func (g *Game) EnemyScheme(enemy *Card, isVillain bool) {
	if enemy.Confused {
		enemy.Confused = false
		g.Logf("%s is confused: its scheme is replaced by discarding the confuse.", enemy.Name())
		return
	}
	n := enemy.Face().SCH + g.attachedSCH(enemy)
	if isVillain {
		n += g.flipBoost(g.drawEncounter())
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

// PlaceThreatWindowed places threat, first offering interrupts such as Great
// Responsibility. The window is scheduled as a step, so effects that follow
// it in the same list resolve after the player answers.
func (g *Game) PlaceThreatWindowed(scheme *Card, n int) {
	if n <= 0 {
		return
	}
	ev := &Event{Trigger: TrigThreatWouldBePlaced, Target: scheme, Amount: n}
	g.Do(func() {
		g.playerWindow(ev, fmt.Sprintf("%d threat is about to be placed on %s (now %d/%d).", n, scheme.Name(), scheme.Threat, scheme.Face().TargetThreat), func() {
			g.PlaceThreat(scheme, ev.Amount)
		})
	})
}
