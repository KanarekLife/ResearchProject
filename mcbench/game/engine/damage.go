package engine

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

func (g *Game) HealVillain(n int) int {
	v := g.S.Villain
	healed := min(n, v.Damage)
	v.Damage -= healed
	g.Logf("%s heals %d damage.", v.Name(), healed)
	return healed
}
