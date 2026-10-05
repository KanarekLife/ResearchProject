package constants

// Metric keys are facts about a finished game, produced by the engine and
// read by the scoring and reporting code. Lower-case to match JSON records.
const (
	MetricWon            = "won"
	MetricLost           = "lost"
	MetricOver           = "over"
	MetricRound          = "round"
	MetricDecisions      = "decisions"
	MetricHeroHP         = "hero_hp"
	MetricHeroMaxHP      = "hero_max_hp"
	MetricVillainDamage  = "villain_damage"
	MetricVillainTotalHP = "villain_total_hp"
	MetricVillainHP      = "villain_hp"
	MetricVillainHPTotal = "villain_hp_total"
	MetricMainThreat     = "main_threat"
	MetricMainTarget     = "main_target"
	MetricSideSchemes    = "side_schemes"
	MetricSideThreat     = "side_threat"
	MetricMinions        = "minions"
	MetricHand           = "hand"
	MetricAllies         = "allies"
)

// Criteria are the scored, normalized measurements derived from the metrics.
const (
	CriterionWin           = "win"
	CriterionVillainDamage = "villain_damage"
	CriterionHeroHP        = "hero_hp"
	CriterionThreat        = "threat"
	CriterionSpeed         = "speed"
)
