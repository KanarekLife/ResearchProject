package engine

import (
	"fmt"

	"mcbench/constants"
)

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

	opts := g.defendOptions(atk)
	resolve := func() { g.resolveAttack(atk, isVillain, onDamaged) }
	if len(opts) == 0 {
		g.Do(resolve)
		return
	}
	opts = append(opts, Option{Key: constants.PrefixDefend + "none", Text: "Do not defend (take the damage)"})
	g.Ask(&Decision{Kind: constants.KindDefend, Prompt: g.defendPrompt(enemy, isVillain), Options: opts})
	g.Do(resolve)
}

// defendOptions lists the ready characters that may defend the attack.
func (g *Game) defendOptions(atk *AttackState) []Option {
	var opts []Option
	if h := g.S.Hero; !h.Exhausted {
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
			Text: fmt.Sprintf("Defend with ally %s (exhaust; it takes all the damage, %d HP left)", g.Label(ally), ally.RemainingHP()),
			do:   func() { ally.Exhausted = true; atk.defender = ally },
		})
	}
	return opts
}

func (g *Game) defendPrompt(enemy *Card, isVillain bool) string {
	prompt := fmt.Sprintf("%s attacks you (base ATK %d", enemy.Name(), enemy.Face().ATK+g.attachedATK(enemy))
	if isVillain {
		prompt += ", plus a face-down boost card"
	}
	return prompt + "). Declare a defender."
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

	finish := func() { g.fireForced(&Event{Trigger: TrigAttackEnded, Source: enemy}) }
	toHero := g.defenderDamage(atk, amount, onDamaged)
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

// defenderDamage applies the attack to the declared defender and returns the
// damage that carries over to the hero.
func (g *Game) defenderDamage(atk *AttackState, amount int, onDamaged func(*Card)) int {
	s := g.S
	switch d := atk.defender; {
	case d == nil:
		return amount
	case d == s.Hero:
		g.Logf("%s defends, reducing the damage by %d.", s.Hero.Name(), g.HeroDEF())
		return max(0, amount-g.HeroDEF())
	default:
		dealt, excess := g.DamageAlly(d, amount)
		if dealt > 0 && onDamaged != nil && contains(s.Play, d) {
			onDamaged(d)
		}
		if atk.Overkill && excess > 0 {
			g.Logf("Overkill: %d excess damage goes to %s.", excess, s.Hero.Name())
			return excess
		}
		return 0
	}
}
