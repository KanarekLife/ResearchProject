package bench

// Criteria are the per-game measurements, each in [0, 1] (higher is better).
//
//	win             1 if the villain was defeated, else 0
//	villain_damage  share of the villain's total HP (all stages) removed
//	hero_hp         share of the hero's HP left at the end (0 if defeated)
//	threat          1 - main-scheme threat / threshold at the end (0 if lost to threat)
//	speed           for wins: 1 - (rounds - 1) / max_rounds; 0 otherwise
//
// The game score is their weighted mean. The weights are part of the
// benchmark definition: change them only deliberately, and never between
// runs you intend to compare.
var DefaultWeights = map[string]float64{
	"win":            0.40,
	"villain_damage": 0.25,
	"hero_hp":        0.10,
	"threat":         0.10,
	"speed":          0.15,
}

// CriteriaOrder is the display order.
var CriteriaOrder = []string{"win", "villain_damage", "hero_hp", "threat", "speed"}

func clamp01(x float64) float64 { return max(0, min(1, x)) }

func ratio(a, b float64) float64 {
	if b <= 0 {
		return 0
	}
	return a / b
}

// Criteria computes the criteria from the engine's final metrics.
func Criteria(m map[string]float64, maxRounds int) map[string]float64 {
	won := m["won"] == 1
	c := map[string]float64{
		"win":            m["won"],
		"villain_damage": clamp01(ratio(m["villain_damage"], m["villain_total_hp"])),
		"hero_hp":        clamp01(ratio(m["hero_hp"], m["hero_max_hp"])),
		"threat":         clamp01(1 - ratio(m["main_threat"], m["main_target"])),
	}
	if won {
		c["speed"] = clamp01(1 - (m["round"]-1)/float64(maxRounds))
		c["villain_damage"] = 1
	} else {
		c["speed"] = 0
	}
	return c
}

// Score is the weighted mean of the criteria. overrides replace default
// weights for the keys they name.
func Score(c map[string]float64, overrides map[string]float64) float64 {
	var sum, total float64
	for k, w := range DefaultWeights {
		if o, ok := overrides[k]; ok {
			w = o
		}
		sum += w * c[k]
		total += w
	}
	return ratio(sum, total)
}
