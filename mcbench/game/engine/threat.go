package engine

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

// CrisisActive reports whether a crisis icon stops player cards removing
// threat from the main scheme. Encounter cards are not affected, but none of
// them removes threat, so RemoveThreat always comes from a player card.
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
		g.leavePlay(s)
		g.S.EncDiscard = append(g.S.EncDiscard, s)
	}
}
