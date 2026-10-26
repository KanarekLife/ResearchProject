package engine

// PlaceThreat adds threat to a scheme. A main scheme that reaches its
// target threat is completed: it advances to its next stage, or the
// players lose after the final one.
func (g *Game) PlaceThreat(s *Card, n int) {
	if n <= 0 {
		return
	}
	s.Threat += n
	g.Logf("%d threat placed on %s (now %d).", n, s.Name(), s.Threat)
	if s != g.S.MainScheme || s.Threat < s.Face().TargetThreat {
		return
	}
	if len(g.S.MainStages) == 0 {
		g.lose(s.Name() + " reached its threat threshold")
		return
	}
	g.advanceMainScheme()
}

// advanceMainScheme replaces the completed main scheme stage with the next
// one: excess threat is lost, the new stage gets its starting threat and its
// When Revealed ability resolves.
func (g *Game) advanceMainScheme() {
	s := g.S
	main := s.MainScheme
	s.MainCompleted += main.Face().TargetThreat
	g.discardAttachments(main)
	next := s.MainStages[0]
	s.MainStages = s.MainStages[1:]
	main.Def, main.Threat = next, next.StartingThreat
	g.Logf("The main scheme stage is completed. It advances to %s (threshold %d).", next.Name, next.TargetThreat)
	if next.Script != nil && next.Script.OnReveal != nil {
		next.Script.OnReveal(g, main)
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
	if s.Face().NoThreatRemoval {
		g.Logf("Threat cannot be removed from %s.", s.Name())
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
