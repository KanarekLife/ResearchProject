package engine

import "fmt"

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

// DamageEnemy deals damage to the villain or a minion and handles defeat. It
// reports whether the enemy (or the villain's stage) was defeated and the
// damage beyond its remaining hit points.
func (g *Game) DamageEnemy(target *Card, amount int) (defeated bool, excess int) {
	if amount <= 0 {
		return false, 0
	}
	if target.Tough {
		target.Tough = false
		g.Logf("%s's tough status prevents the damage and is discarded.", target.Name())
		return false, 0
	}
	ev := &Event{Trigger: TrigEnemyWouldTakeDamage, Target: target, Amount: amount}
	g.fireForced(ev)
	if ev.Amount <= 0 {
		return false, 0
	}
	excess = max(0, ev.Amount-g.RemainingHP(target))
	target.Damage += ev.Amount
	g.Logf("%s takes %d damage (%d HP left).", target.Name(), ev.Amount, max(0, g.RemainingHP(target)))
	defeated = g.RemainingHP(target) <= 0
	if defeated {
		g.defeatEnemy(target)
	} else {
		excess = 0
	}
	// "After it takes damage" resolves after a defeat, so it reaches the
	// villain's next stage but not a defeated minion.
	if !g.result.Over && g.isInPlay(target) {
		g.fireForced(&Event{Trigger: TrigEnemyDamaged, Target: target, Amount: ev.Amount})
	}
	return defeated, excess
}

func (g *Game) defeatEnemy(target *Card) {
	if target == g.S.Villain {
		g.defeatVillainStage()
	} else {
		g.defeatMinion(target)
	}
}

// checkDefeats defeats each character left with no hit points, for when a
// hit point modifier ends rather than damage being dealt.
func (g *Game) checkDefeats() {
	s := g.S
	if g.RemainingHP(s.Hero) <= 0 {
		g.lose(s.Hero.Name() + " was defeated")
		return
	}
	if g.RemainingHP(s.Villain) <= 0 {
		g.Logf("%s has no hit points left.", s.Villain.Name())
		g.defeatVillainStage()
	}
	for _, m := range append([]*Card(nil), s.Minions...) {
		if contains(s.Minions, m) && g.RemainingHP(m) <= 0 {
			g.defeatMinion(m)
		}
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
	ev := &Event{Trigger: TrigMinionDefeated, Target: m}
	g.fireForced(ev)
	g.Logf("%s is defeated.", m.Name())
	g.S.Minions = remove(g.S.Minions, m)
	drone := m.Drone != nil
	g.leavePlay(m)
	if drone {
		// A facedown Drone is its owner's card (Ultron Drones).
		g.S.Discard = append(g.S.Discard, m)
		g.Logf("The Drone was %s; it goes to your discard pile.", m.Def.Name)
	} else {
		g.S.EncDiscard = append(g.S.EncDiscard, m)
	}
	g.Do(func() { g.playerWindow(ev, m.Name()+" was defeated.", func() {}) })
}

// leavePlay resets a card that moved out of play: it keeps no memory of its
// previous state, and each card attached to it is discarded.
func (g *Game) leavePlay(c *Card) {
	g.discardAttachments(c)
	c.Exhausted, c.Damage, c.Threat, c.Counters = false, 0, 0, 0
	c.Stunned, c.Confused, c.Tough, c.Facedown = false, false, false, false
	c.Drone, c.TempATK, c.TempTHW = nil, 0, 0
}

func (g *Game) discardAttachments(c *Card) {
	for _, a := range c.Attached {
		g.leavePlay(a)
		if a.Def.IsPlayerCard() {
			g.S.Discard = append(g.S.Discard, a)
		} else {
			g.S.EncDiscard = append(g.S.EncDiscard, a)
		}
	}
	c.Attached = nil
}

// Detach discards a card in play (a card that already left play stays
// where it is).
func (g *Game) Detach(a *Card) {
	if !g.isInPlay(a) {
		return
	}
	g.inPlay(func(c *Card) {
		if contains(c.Attached, a) {
			c.Attached = remove(c.Attached, a)
		}
	})
	g.S.Play = remove(g.S.Play, a)
	g.leavePlay(a)
	if a.Def.IsPlayerCard() {
		g.S.Discard = append(g.S.Discard, a)
	} else {
		g.S.EncDiscard = append(g.S.EncDiscard, a)
	}
}

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
	g.Logf("%s takes %d damage (%d HP left).", h.Name(), amount, max(0, g.RemainingHP(h)))
	if g.RemainingHP(h) <= 0 {
		g.lose(h.Name() + " was defeated")
	}
	return amount
}

// HarmHero deals damage to your hero from something other than an attack,
// first offering interrupts to it (such as Cosmic Flight). Without any, the
// damage is dealt at once. after, if set, gets the damage dealt.
func (g *Game) HarmHero(amount int, after func(dealt int)) {
	if amount <= 0 {
		return
	}
	done := func(n int) {
		if after != nil {
			after(n)
		}
	}
	ev := &Event{Trigger: TrigHeroWouldTakeDamage, Target: g.S.Hero, Amount: amount}
	prompt := fmt.Sprintf("You are about to take %d damage.", amount)
	if g.S.Hero.Tough || len(g.windowOptions(ev, prompt, func() {})) == 0 {
		done(g.DamageHero(amount))
		return
	}
	g.playerWindow(ev, prompt, func() { done(g.DamageHero(ev.Amount)) })
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
	rem := g.RemainingHP(a)
	if amount > rem {
		excess = amount - rem
	}
	a.Damage += amount
	g.Logf("%s takes %d damage.", a.Name(), amount)
	if g.RemainingHP(a) <= 0 {
		g.Logf("%s is defeated.", a.Name())
		g.S.Play = remove(g.S.Play, a)
		g.leavePlay(a)
		g.S.Discard = append(g.S.Discard, a)
	}
	return amount, excess
}

func (g *Game) HealHero(n int) {
	g.Heal(g.S.Hero, n)
}

// Heal removes up to n damage from a character.
func (g *Game) Heal(c *Card, n int) {
	healed := min(n, c.Damage)
	c.Damage -= healed
	g.Logf("%s heals %d damage (%d HP left).", g.Label(c), healed, g.RemainingHP(c))
}

func (g *Game) HealVillain(n int) int {
	v := g.S.Villain
	healed := min(n, v.Damage)
	v.Damage -= healed
	g.Logf("%s heals %d damage.", v.Name(), healed)
	return healed
}
