package benchmark

import (
	"testing"

	"mcbench/constants"
	"mcbench/game/session"
)

// The identity is told apart from allies by the change-form options, so a
// hero attack, a hero defence, an unused ally and a risky alter-ego turn end
// are each counted as such.
func TestBehaviourClassifiesTurns(t *testing.T) {
	turn := func(key, form string, threat int, offered ...string) session.Choice {
		return session.Choice{Kind: constants.KindTurn, Key: key, Form: form, Threat: threat, Offered: offered}
	}
	r := Record{Status: constants.Lost, Rounds: 2,
		Stats: map[string]float64{constants.MetricHeroHP: 4, constants.MetricMainTarget: 7},
		Choices: []session.Choice{
			turn("attack:Spider-Man>Rhino", constants.TypeHero, 1, "attack:Spider-Man>Rhino", "change_form:Peter Parker"),
			turn("attack:Black Cat>Rhino", constants.TypeHero, 1),
			{Kind: constants.KindDefend, Key: "defend:Spider-Man"},
			turn(constants.KeyEndTurn, constants.TypeAlterEgo, 4, "thwart:Black Cat>The Break-In!", "change_form:Spider-Man"),
		}}
	var b Behaviour
	b.add(r)
	if b.Scheme != 1 || b.Basic != 1 || b.Attacks != 1 || b.HeroDefends != 1 || b.Actions != 2 {
		t.Errorf("play measures %+v", b)
	}
	if b.Ends != 1 || b.AERisky != 1 || b.IdleAlly != 1 || b.IdleHero != 0 || !b.Recorded {
		t.Errorf("turn-end measures %+v", b)
	}
}
