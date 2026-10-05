package benchmark

import (
	"math"
	"testing"

	"mcbench/constants"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// The weights are the benchmark definition: they must sum to 1 so a score is
// directly a share of the maximum, and every criterion must be weighted.
func TestDefaultWeightsSumToOne(t *testing.T) {
	var sum float64
	for _, k := range CriteriaOrder {
		w, ok := DefaultWeights[k]
		if !ok {
			t.Fatalf("criterion %q has no default weight", k)
		}
		sum += w
	}
	if !near(sum, 1) {
		t.Fatalf("default weights sum to %v, want 1", sum)
	}
}

func TestCriteria(t *testing.T) {
	const maxRounds = 20
	tests := []struct {
		name  string
		stats map[string]float64
		want  map[string]float64
	}{
		{
			name: "win in round 1 is a perfect speed score",
			stats: map[string]float64{
				constants.MetricWon: 1, constants.MetricRound: 1,
				constants.MetricHeroHP: 5, constants.MetricHeroMaxHP: 10,
				constants.MetricMainThreat: 0, constants.MetricMainTarget: 10,
			},
			want: map[string]float64{
				constants.CriterionWin: 1, constants.CriterionVillainDamage: 1,
				constants.CriterionHeroHP: 0.5, constants.CriterionThreat: 1,
				constants.CriterionSpeed: 1,
			},
		},
		{
			name: "loss scores no speed and partial villain damage",
			stats: map[string]float64{
				constants.MetricRound:         7,
				constants.MetricVillainDamage: 6, constants.MetricVillainTotalHP: 24,
				constants.MetricMainThreat: 8, constants.MetricMainTarget: 10,
			},
			want: map[string]float64{
				constants.CriterionWin: 0, constants.CriterionVillainDamage: 0.25,
				constants.CriterionHeroHP: 0, constants.CriterionThreat: 0.2,
				constants.CriterionSpeed: 0,
			},
		},
		{
			name: "threat past the target clamps to zero, not negative",
			stats: map[string]float64{
				constants.MetricMainThreat: 15, constants.MetricMainTarget: 10,
			},
			want: map[string]float64{constants.CriterionThreat: 0},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Criteria(tc.stats, maxRounds)
			for k, w := range tc.want {
				if !near(got[k], w) {
					t.Errorf("%s = %v, want %v", k, got[k], w)
				}
			}
		})
	}
}

func TestScore(t *testing.T) {
	perfect := map[string]float64{}
	for _, k := range CriteriaOrder {
		perfect[k] = 1
	}
	if got := Score(perfect, nil); !near(got, 1) {
		t.Errorf("perfect game with default weights = %v, want 1", got)
	}
	if got := Score(map[string]float64{}, nil); got != 0 {
		t.Errorf("empty criteria = %v, want 0", got)
	}

	// Weights are normalised, so only their ratio matters.
	only := Weights{constants.CriterionWin: 5}
	if got := Score(map[string]float64{constants.CriterionWin: 1}, only); !near(got, 1) {
		t.Errorf("single-criterion weights = %v, want 1", got)
	}
}
