package engine

import "fmt"

// TrigAttackEnded fires on cards in play after an enemy attack resolves.
const TrigAttackEnded Trigger = "attack_ended"

// Label names a card uniquely for keys and option text: "Name", or
// "Name#ID" when another card in play shares the name.
func (g *Game) Label(c *Card) string {
	n := 0
	g.inPlay(func(x *Card) {
		if x.Name() == c.Name() {
			n++
		}
	})
	if n > 1 {
		return fmt.Sprintf("%s#%d", c.Name(), c.ID)
	}
	return c.Name()
}

// --- statuses ---------------------------------------------------------------

func (g *Game) Stun(c *Card) {
	if !c.Stunned {
		c.Stunned = true
		g.Logf("%s is stunned.", c.Name())
	}
}

func (g *Game) Confuse(c *Card) {
	if !c.Confused {
		c.Confused = true
		g.Logf("%s is confused.", c.Name())
	}
}

func (g *Game) GiveTough(c *Card) {
	if !c.Tough {
		c.Tough = true
		g.Logf("%s gains a tough status.", c.Name())
	}
}

// --- damage and healing -------------------------------------------------------

// DamageEnemy deals damage to the villain or a minion and handles defeat.
func (g *Game) DamageEnemy(target *Card, amount int) {
	if amount <= 0 {
		return
	}
	if target.Tough {
		target.Tough = false
		g.Logf("%s's tough status prevents the damage and is discarded.", target.Name())
		return
	}
	ev := &Event{Trigger: TrigEnemyWouldTakeDamage, Target: target, Amount: amount}
	g.fireForced(ev)
	if ev.Amount <= 0 {
		return
	}
	target.Damage += ev.Amount
	g.Logf("%s takes %d damage (%d HP left).", target.Name(), ev.Amount, max(0, target.RemainingHP()))
	if target.RemainingHP() > 0 {
		return
	}
	if target == g.S.Villain {
		g.defeatVillainStage()
	} else {
		g.defeatMinion(target)
	}
}

func (g *Game) defeatVillainStage() {
	v := g.S.Villain
	if len(g.S.VillainStages) == 0 {
		g.Logf("%s is defeated.", v.Name())
		g.win(v.Name() + " defeated")
		return
	}
	next := g.S.VillainStages[0]
	g.S.VillainStages = g.S.VillainStages[1:]
	v.Def = next
	v.Damage = 0
	g.Logf("Villain stage defeated. %s advances to stage %s (%d HP).", next.Name, next.Code, next.HP)
	if next.Toughness {
		g.GiveTough(v)
	}
	if next.Script != nil && next.Script.OnReveal != nil {
		next.Script.OnReveal(g, v)
	}
}

func (g *Game) defeatMinion(m *Card) {
	g.fireForced(&Event{Trigger: TrigMinionDefeated, Target: m})
	g.Logf("%s is defeated.", m.Name())
	g.S.Minions = remove(g.S.Minions, m)
	g.discardAttachments(m)
	m.Damage, m.Stunned, m.Confused, m.Tough = 0, false, false, false
	g.S.EncDiscard = append(g.S.EncDiscard, m)
}

func (g *Game) discardAttachments(c *Card) {
	for _, a := range c.Attached {
		if a.Def.IsPlayerCard() {
			g.S.Discard = append(g.S.Discard, a)
		} else {
			g.S.EncDiscard = append(g.S.EncDiscard, a)
		}
	}
	c.Attached = nil
}

// Detach moves an attachment from whatever it is attached to into its discard.
func (g *Game) Detach(a *Card) {
	g.inPlay(func(c *Card) {
		if contains(c.Attached, a) {
			c.Attached = remove(c.Attached, a)
		}
	})
	g.S.Play = remove(g.S.Play, a)
	if a.Def.IsPlayerCard() {
		g.S.Discard = append(g.S.Discard, a)
	} else {
		g.S.EncDiscard = append(g.S.EncDiscard, a)
	}
}

// DamageHero deals damage to the identity and returns the damage dealt.
func (g *Game) DamageHero(amount int) int {
	h := g.S.Hero
	if amount <= 0 {
		return 0
	}
	if h.Tough {
		h.Tough = false
		g.Logf("%s's tough status prevents the damage and is discarded.", h.Name())
		return 0
	}
	h.Damage += amount
	g.Logf("%s takes %d damage (%d HP left).", h.Name(), amount, max(0, h.RemainingHP()))
	if h.RemainingHP() <= 0 {
		g.lose(h.Name() + " was defeated")
	}
	return amount
}

// DamageAlly deals damage to an ally and returns the excess beyond its HP.
func (g *Game) DamageAlly(a *Card, amount int) (dealt, excess int) {
	if amount <= 0 {
		return 0, 0
	}
	if a.Tough {
		a.Tough = false
		g.Logf("%s's tough status prevents the damage and is discarded.", a.Name())
		return 0, 0
	}
	rem := a.RemainingHP()
	if amount > rem {
		excess = amount - rem
	}
	a.Damage += amount
	g.Logf("%s takes %d damage.", a.Name(), amount)
	if a.RemainingHP() <= 0 {
		g.Logf("%s is defeated.", a.Name())
		g.S.Play = remove(g.S.Play, a)
		g.discardAttachments(a)
		a.Damage, a.Exhausted, a.Stunned, a.Confused, a.Tough = 0, false, false, false, false
		g.S.Discard = append(g.S.Discard, a)
	}
	return amount, excess
}

func (g *Game) HealHero(n int) {
	h := g.S.Hero
	healed := min(n, h.Damage)
	h.Damage -= healed
	g.Logf("%s heals %d damage (%d HP left).", h.Name(), healed, h.RemainingHP())
}

// HealVillain heals the villain and returns the amount healed.
func (g *Game) HealVillain(n int) int {
	v := g.S.Villain
	healed := min(n, v.Damage)
	v.Damage -= healed
	g.Logf("%s heals %d damage.", v.Name(), healed)
	return healed
}

// --- threat ---------------------------------------------------------------------

// PlaceThreat adds threat to a scheme; reaching the main scheme's target loses.
func (g *Game) PlaceThreat(s *Card, n int) {
	if n <= 0 {
		return
	}
	s.Threat += n
	g.Logf("%d threat placed on %s (now %d).", n, s.Name(), s.Threat)
	if s == g.S.MainScheme && s.Threat >= s.Face().TargetThreat {
		g.lose(s.Name() + " reached its threat threshold")
	}
}

// CrisisActive reports whether a crisis icon blocks main-scheme thwarting.
func (g *Game) CrisisActive() bool {
	for _, s := range g.S.SideSchemes {
		if s.Face().Crisis {
			return true
		}
	}
	return false
}

// RemoveThreat removes threat from a scheme; an emptied side scheme is defeated.
func (g *Game) RemoveThreat(s *Card, n int) {
	if s == g.S.MainScheme && g.CrisisActive() {
		g.Logf("A crisis icon prevents removing threat from %s.", s.Name())
		return
	}
	removed := min(n, s.Threat)
	s.Threat -= removed
	g.Logf("%d threat removed from %s (now %d).", removed, s.Name(), s.Threat)
	if s != g.S.MainScheme && s.Threat == 0 && contains(g.S.SideSchemes, s) {
		g.Logf("%s is defeated.", s.Name())
		g.S.SideSchemes = remove(g.S.SideSchemes, s)
		if sc := s.Def.Script; sc != nil && sc.OnDefeated != nil {
			sc.OnDefeated(g, s)
		}
		g.discardAttachments(s)
		g.S.EncDiscard = append(g.S.EncDiscard, s)
	}
}

// --- player attacks and thwarts --------------------------------------------------

// Attack performs an attack by a player character, honoring stun.
func (g *Game) Attack(attacker, target *Card, amount, consequential int) {
	if attacker.Stunned {
		attacker.Stunned = false
		g.Logf("%s is stunned: the attack is replaced by discarding the stun.", attacker.Name())
		return
	}
	g.Logf("%s attacks %s for %d.", attacker.Name(), target.Name(), amount)
	g.DamageEnemy(target, amount)
	if consequential > 0 && contains(g.S.Play, attacker) {
		g.DamageAlly(attacker, consequential)
	}
}

// Thwart performs a thwart by a player character, honoring confuse.
func (g *Game) Thwart(thwarter, scheme *Card, amount, consequential int) {
	if thwarter.Confused {
		thwarter.Confused = false
		g.Logf("%s is confused: the thwart is replaced by discarding the confuse.", thwarter.Name())
		return
	}
	g.Logf("%s thwarts %s for %d.", thwarter.Name(), scheme.Name(), amount)
	g.RemoveThreat(scheme, amount)
	if consequential > 0 && contains(g.S.Play, thwarter) {
		g.DamageAlly(thwarter, consequential)
	}
	g.fireForced(&Event{Trigger: TrigThwarted, Source: thwarter, Target: scheme})
}

// Enemies lists the enemies a player attack may target (guard enforced).
func (g *Game) Enemies() []*Card {
	var guards []*Card
	for _, m := range g.S.Minions {
		if m.Face().Guard {
			guards = append(guards, m)
		}
	}
	if len(guards) > 0 {
		return append([]*Card(nil), g.S.Minions...)
	}
	return append([]*Card{g.S.Villain}, g.S.Minions...)
}

// AllEnemies lists the villain and every minion, ignoring guard.
func (g *Game) AllEnemies() []*Card {
	return append([]*Card{g.S.Villain}, g.S.Minions...)
}

// ThwartableSchemes lists schemes with threat that may be thwarted.
func (g *Game) ThwartableSchemes() []*Card {
	var out []*Card
	if g.S.MainScheme.Threat > 0 && !g.CrisisActive() {
		out = append(out, g.S.MainScheme)
	}
	for _, s := range g.S.SideSchemes {
		if s.Threat > 0 {
			out = append(out, s)
		}
	}
	return out
}

// HostOf returns the card c is attached to, or nil.
func (g *Game) HostOf(c *Card) *Card {
	var host *Card
	g.inPlay(func(x *Card) {
		if contains(x.Attached, c) {
			host = x
		}
	})
	return host
}

// ChooseScheme asks the player to pick a scheme with threat to remove n
// threat from (used by effects that say "remove N threat from a scheme").
func (g *Game) ChooseScheme(source string, n int) {
	g.Do(func() {
		schemes := g.ThwartableSchemes()
		if len(schemes) == 0 {
			return
		}
		var opts []Option
		for _, sc := range schemes {
			scheme := sc
			opts = append(opts, NewAction(
				"effect:"+source+">"+g.Label(scheme),
				fmt.Sprintf("Remove %d threat from %s", n, g.Label(scheme)),
				func() { g.RemoveThreat(scheme, n) },
			))
		}
		g.Ask(&Decision{Kind: "choice", Prompt: fmt.Sprintf("%s: choose a scheme to remove %d threat from.", source, n), Options: opts})
	})
}

// ChooseEnemy asks the player to pick an enemy (guard ignored) for an effect.
func (g *Game) ChooseEnemy(source, verb string, apply func(*Card)) {
	g.Do(func() {
		var opts []Option
		for _, e := range g.AllEnemies() {
			enemy := e
			opts = append(opts, NewAction(
				"effect:"+source+">"+g.Label(enemy),
				fmt.Sprintf("%s %s", verb, g.Label(enemy)),
				func() { apply(enemy) },
			))
		}
		g.Ask(&Decision{Kind: "choice", Prompt: fmt.Sprintf("%s: choose an enemy.", source), Options: opts})
	})
}
