package engine

import "mcbench/constants"

// Metrics are numeric facts about the current state, used for grading.
func (g *Game) Metrics() map[string]float64 {
	s := g.S
	villainHP := g.villainHP()
	if g.result.Over && g.result.Won {
		villainHP = 0
	}
	return map[string]float64{
		constants.MetricVillainDamage:  float64(max(0, s.VillainTotalHP-villainHP)),
		constants.MetricVillainTotalHP: float64(s.VillainTotalHP),
		constants.MetricHeroMaxHP:      float64(s.Hero.Face().HP),
		constants.MetricMainTarget:     float64(s.MainScheme.Face().TargetThreat),
		constants.MetricWon:            boolFloat(g.result.Over && g.result.Won),
		constants.MetricLost:           boolFloat(g.result.Over && !g.result.Won),
		constants.MetricOver:           boolFloat(g.result.Over),
		constants.MetricRound:          float64(s.Round),
		constants.MetricHeroHP:         float64(max(0, s.Hero.RemainingHP())),
		constants.MetricVillainHPTotal: float64(max(0, villainHP)),
		constants.MetricVillainHP:      float64(max(0, s.Villain.RemainingHP())),
		constants.MetricMainThreat:     float64(s.MainScheme.Threat),
		constants.MetricSideSchemes:    float64(len(s.SideSchemes)),
		constants.MetricSideThreat:     float64(sumThreat(s.SideSchemes)),
		constants.MetricMinions:        float64(len(s.Minions)),
		constants.MetricHand:           float64(len(s.Hand)),
		constants.MetricAllies:         float64(g.allyCount()),
		constants.MetricDecisions:      float64(g.Decisions),
	}
}

// villainHP is the villain's remaining HP over all stages.
func (g *Game) villainHP() int {
	hp := g.S.Villain.RemainingHP()
	for _, st := range g.S.VillainStages {
		hp += st.HP
	}
	return hp
}

func boolFloat(v bool) float64 {
	if v {
		return 1
	}
	return 0
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
