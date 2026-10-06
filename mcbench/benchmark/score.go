package benchmark

import "mcbench/constants"

// Criteria are the per-game measurements, each in [0, 1] (higher is better).
//
//	win             1 if the villain was defeated, else 0
//	villain_damage  share of the villain's total HP (all stages) removed
//	hero_hp         share of the hero's HP left at the end (0 if defeated)
//	threat          1 - main-scheme threat / threshold at the end
//	speed           for wins: 1 - (rounds - 1) / max_rounds; 0 otherwise
//
// The game score is their weighted mean. The weights are the benchmark
// definition: change them only deliberately, and never between runs you
// intend to compare.
var (
	// DefaultWeights are used when a run supplies none.
	DefaultWeights = Weights{
		constants.CriterionWin:           0.40,
		constants.CriterionVillainDamage: 0.25,
		constants.CriterionHeroHP:        0.10,
		constants.CriterionThreat:        0.10,
		constants.CriterionSpeed:         0.15,
	}
	// CriteriaOrder is the display order.
	CriteriaOrder = []string{
		constants.CriterionWin, constants.CriterionVillainDamage,
		constants.CriterionHeroHP, constants.CriterionThreat, constants.CriterionSpeed,
	}
)

// Weights map a criterion to its share of the score.
type Weights map[string]float64

// Criteria computes the criteria from the engine's final metrics.
func Criteria(stats map[string]float64, maxRounds int) map[string]float64 {
	won := stats[constants.MetricWon] == 1
	c := map[string]float64{
		constants.CriterionWin: stats[constants.MetricWon],
		constants.CriterionVillainDamage: clamp01(ratio(
			stats[constants.MetricVillainDamage], stats[constants.MetricVillainTotalHP])),
		constants.CriterionHeroHP: clamp01(ratio(
			stats[constants.MetricHeroHP], stats[constants.MetricHeroMaxHP])),
		constants.CriterionThreat: clamp01(1 - ratio(
			stats[constants.MetricMainThreat], stats[constants.MetricMainTarget])),
	}
	if won {
		c[constants.CriterionSpeed] = clamp01(1 - (stats[constants.MetricRound]-1)/float64(maxRounds))
		c[constants.CriterionVillainDamage] = 1
	} else {
		c[constants.CriterionSpeed] = 0
	}
	return c
}

// Score is the weighted mean of the criteria. An empty weights set uses
// DefaultWeights.
func Score(criteria map[string]float64, weights Weights) float64 {
	if len(weights) == 0 {
		weights = DefaultWeights
	}
	var sum, total float64
	for _, k := range CriteriaOrder {
		w := weights[k]
		sum += w * criteria[k]
		total += w
	}
	return ratio(sum, total)
}

func clamp01(x float64) float64 { return max(0, min(1, x)) }

func ratio(a, b float64) float64 {
	if b <= 0 {
		return 0
	}
	return a / b
}
