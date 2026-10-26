package engine

import (
	"fmt"

	"mcbench/constants"
)

// AttackState is shared across the steps of one enemy attack.
type AttackState struct {
	Enemy    *Card
	Overkill bool
	boosts   []*Card // the villain's facedown boost cards
	defender *Card   // nil = undefended; hero or an ally
}

// EnemyAttack resolves an enemy attack against the hero. onDamaged is
// called with each character damaged by the attack and the damage dealt.
func (g *Game) EnemyAttack(enemy *Card, isVillain bool, onDamaged func(*Card, int)) {
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
	if !isVillain {
		g.declareDefense(atk, isVillain, onDamaged)
		return
	}
	ev = &Event{Trigger: TrigVillainAttacks, Source: enemy}
	g.fireForced(ev)
	g.playerWindow(ev, fmt.Sprintf("%s initiates an attack against you.", enemy.Name()), func() {
		// Dealt before the defender is declared.
		for i := 0; i <= enemy.Face().ExtraBoost; i++ {
			if c := g.drawEncounter(); c != nil {
				atk.boosts = append(atk.boosts, c)
			}
		}
		g.declareDefense(atk, isVillain, onDamaged)
	})
}

// declareDefense asks for a defender, then resolves the attack.
func (g *Game) declareDefense(atk *AttackState, isVillain bool, onDamaged func(*Card, int)) {
	enemy := atk.Enemy
	g.Logf("%s attacks %s.", enemy.Name(), g.S.Hero.Name())
	opts := g.defendOptions(atk)
	resolve := func() { g.resolveAttack(atk, isVillain, onDamaged) }
	if len(opts) == 0 {
		g.Do(resolve)
		return
	}
	if !g.mustDefendWithAlly(enemy) {
		opts = append(opts, Option{Key: constants.PrefixDefend + "none", Text: "Do not defend (take the damage)"})
	}
	g.Ask(&Decision{Kind: constants.KindDefend, Prompt: g.defendPrompt(enemy, isVillain), Options: opts})
	g.Do(resolve)
}

// mustDefendWithAlly reports whether the enemy's attack must be defended by
// a ready ally, because it says so and you have one.
func (g *Game) mustDefendWithAlly(enemy *Card) bool {
	if !enemy.Face().AllyMustDefend {
		return false
	}
	for _, a := range g.S.Play {
		if a.Def.Type == TypeAlly && !a.Exhausted {
			return true
		}
	}
	return false
}

// defendOptions lists the ready characters that may defend the attack.
func (g *Game) defendOptions(atk *AttackState) []Option {
	var opts []Option
	if h := g.S.Hero; !h.Exhausted && g.IsHero() && !g.mustDefendWithAlly(atk.Enemy) {
		opts = append(opts, Option{
			Key:  constants.PrefixDefend + h.Name(),
			Text: fmt.Sprintf("Defend with %s (exhaust; DEF %d reduces the damage)", h.Name(), g.HeroDEF()),
			do:   func() { h.Exhausted = true; atk.defender = h },
		})
	}
	for _, a := range g.S.Play {
		if a.Def.Type != TypeAlly || a.Exhausted {
			continue
		}
		ally := a
		opts = append(opts, Option{
			Key:  constants.PrefixDefend + g.Label(ally),
			Text: fmt.Sprintf("Defend with ally %s (exhaust; it takes all the damage, %d HP left)", g.Label(ally), g.RemainingHP(ally)),
			do:   func() { ally.Exhausted = true; atk.defender = ally },
		})
	}
	return opts
}

func (g *Game) defendPrompt(enemy *Card, isVillain bool) string {
	prompt := fmt.Sprintf("%s attacks you (base ATK %d", enemy.Name(), g.EnemyATK(enemy))
	if isVillain {
		if n := 1 + enemy.Face().ExtraBoost; n > 1 {
			prompt += fmt.Sprintf(", plus %d face-down boost cards", n)
		} else {
			prompt += ", plus a face-down boost card"
		}
	}
	if g.mustDefendWithAlly(enemy) {
		prompt += "; it must be defended by an ally"
	}
	return prompt + "). Declare a defender."
}

// resolveAttack flips the boost cards, resolves their boost abilities, then
// deals the damage.
func (g *Game) resolveAttack(atk *AttackState, isVillain bool, onDamaged func(*Card, int)) {
	enemy := atk.Enemy
	for _, a := range enemy.Attached {
		if a.Face().AttachOverkill {
			atk.Overkill = true
		}
	}
	boost := 0
	var steps []func()
	for _, b := range atk.boosts {
		boost += g.flipBoost(b)
		steps = append(steps, func() { g.boostAbility(b, &Event{Source: enemy, Target: atk.defender, Attack: true}) })
	}
	steps = append(steps, func() { g.dealAttack(atk, g.EnemyATK(enemy)+boost, onDamaged) })
	g.Do(steps...)
}

// boostAbility resolves the "Boost" ability a flipped boost card has.
func (g *Game) boostAbility(b *Card, ev *Event) {
	if b != nil && b.Def.Script != nil && b.Def.Script.OnBoost != nil {
		b.Def.Script.OnBoost(g, b, ev)
	}
}

func (g *Game) dealAttack(atk *AttackState, amount int, onDamaged func(*Card, int)) {
	s := g.S
	enemy := atk.Enemy
	onDamaged = g.withBoost(atk, onDamaged)
	g.Logf("%s's attack has %d ATK.", enemy.Name(), amount)

	finish := func() { g.fireForced(&Event{Trigger: TrigAttackEnded, Source: enemy}) }
	toHero := g.defenderDamage(atk, amount, onDamaged)
	if toHero <= 0 {
		g.Do(finish)
		return
	}
	if s.Hero.Tough {
		// A status card resolves before any interrupt to the damage.
		g.DamageHero(toHero)
		g.Do(finish)
		return
	}
	ev := &Event{Trigger: TrigWouldTakeAttackDamage, Source: enemy, Target: s.Hero, Amount: toHero}
	g.playerWindow(ev, fmt.Sprintf("You are about to take %d damage from %s's attack.", toHero, enemy.Name()), func() {
		if n := g.DamageHero(ev.Amount); n > 0 {
			onDamaged(s.Hero, n)
		}
		g.Do(finish)
	})
}

// withBoost extends the attack's onDamaged callback with the boost cards'
// "Boost" abilities for damaged characters and the forced abilities that
// follow an attack damaging a character. Those only affect a character
// still in play.
func (g *Game) withBoost(atk *AttackState, onDamaged func(*Card, int)) func(*Card, int) {
	return func(c *Card, n int) {
		if onDamaged != nil {
			onDamaged(c, n)
		}
		if c != g.S.Hero && !contains(g.S.Play, c) {
			return
		}
		for _, b := range atk.boosts {
			if sc := b.Def.Script; sc != nil && sc.BoostOnDamaged != nil {
				sc.BoostOnDamaged(g, b, c)
			}
		}
		g.fireForced(&Event{Trigger: TrigAttackDamaged, Source: atk.Enemy, Target: c, Amount: n})
	}
}

// defenderDamage applies the attack to the declared defender and returns the
// damage that carries over to the hero.
func (g *Game) defenderDamage(atk *AttackState, amount int, onDamaged func(*Card, int)) int {
	s := g.S
	switch d := atk.defender; {
	case d == nil:
		return amount
	case d == s.Hero:
		g.Logf("%s defends, reducing the damage by %d.", s.Hero.Name(), g.HeroDEF())
		return max(0, amount-g.HeroDEF())
	default:
		dealt, excess := g.DamageAlly(d, amount)
		if dealt > 0 {
			onDamaged(d, dealt)
		}
		if atk.Overkill && excess > 0 {
			g.Logf("Overkill: %d excess damage goes to %s.", excess, s.Hero.Name())
			return excess
		}
		return 0
	}
}
