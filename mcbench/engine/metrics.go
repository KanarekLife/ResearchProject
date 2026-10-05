package engine

// Metrics are numeric facts about the current state, used for grading.
func (g *Game) Metrics() map[string]float64 {
	s := g.S
	b := func(v bool) float64 {
		if v {
			return 1
		}
		return 0
	}
	villainHP := s.Villain.RemainingHP()
	for _, st := range s.VillainStages {
		villainHP += st.HP
	}
	if g.result.Over && g.result.Won {
		villainHP = 0
	}
	villainDamage := s.VillainTotalHP - villainHP
	return map[string]float64{
		"villain_damage":   float64(max(0, villainDamage)),
		"villain_total_hp": float64(s.VillainTotalHP),
		"hero_max_hp":      float64(s.Hero.Face().HP),
		"main_target":      float64(s.MainScheme.Face().TargetThreat),
		"won":              b(g.result.Over && g.result.Won),
		"lost":             b(g.result.Over && !g.result.Won),
		"over":             b(g.result.Over),
		"round":            float64(s.Round),
		"hero_hp":          float64(max(0, s.Hero.RemainingHP())),
		"villain_hp_total": float64(max(0, villainHP)),
		"villain_hp":       float64(max(0, s.Villain.RemainingHP())),
		"main_threat":      float64(s.MainScheme.Threat),
		"side_schemes":     float64(len(s.SideSchemes)),
		"side_threat":      float64(sumThreat(s.SideSchemes)),
		"minions":          float64(len(s.Minions)),
		"hand":             float64(len(s.Hand)),
		"allies":           float64(g.allyCount()),
		"decisions":        float64(g.Decisions),
	}
}

func minionHP(cs []*Card) (n int) {
	for _, c := range cs {
		n += max(0, c.RemainingHP())
	}
	return
}

func sumThreat(cs []*Card) (n int) {
	for _, c := range cs {
		n += c.Threat
	}
	return
}
