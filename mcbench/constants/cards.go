package constants

// Card types (the names used in card YAML).
const (
	TypeHero       = "hero"
	TypeAlterEgo   = "alter-ego"
	TypeAlly       = "ally"
	TypeEvent      = "event"
	TypeSupport    = "support"
	TypeUpgrade    = "upgrade"
	TypeResource   = "resource"
	TypeVillain    = "villain"
	TypeMainScheme = "main_scheme"
	TypeSideScheme = "side_scheme"
	TypeMinion     = "minion"
	TypeTreachery  = "treachery"
	TypeAttachment = "attachment"
	TypeObligation = "obligation"
)

// Resources (the names used in card YAML).
const (
	ResourcePhysical = "physical"
	ResourceMental   = "mental"
	ResourceEnergy   = "energy"
	ResourceWild     = "wild"
)

// Trigger names (the values of Ability.On).
const (
	TrigEnemyWouldAttack      = "enemy_would_attack"
	TrigVillainAttacks        = "villain_attacks"
	TrigWouldTakeAttackDamage = "would_take_attack_damage"
	TrigEnemyWouldTakeDamage  = "enemy_would_take_damage"
	TrigTreacheryRevealed     = "treachery_revealed"
	TrigMinionDefeated        = "minion_defeated"
	TrigVillainSchemes        = "villain_schemes"
	TrigThreatWouldBePlaced   = "threat_would_be_placed"
	TrigThwarted              = "thwarted"
	TrigRoundEnd              = "round_end"
	TrigAttackEnded           = "attack_ended"
)

// Ability triggers (the value of Ability.Trigger).
const (
	AbilityPlay      = "play"
	AbilityAction    = "action"
	AbilityReveal    = "reveal"
	AbilityDefeated  = "defeated"
	AbilityInterrupt = "interrupt"
	AbilityForced    = "forced"
	AbilityResource  = "resource"
	AbilityStat      = "stat"
	AbilityBoost     = "boost"
)

// Effect verbs (the value of Effect.Verb).
const (
	VerbAttack             = "attack"
	VerbDamageEnemy        = "damage_enemy"
	VerbDamageHero         = "damage_hero"
	VerbThwart             = "thwart"
	VerbRemoveThreat       = "remove_threat"
	VerbPlaceThreat        = "place_threat"
	VerbDraw               = "draw"
	VerbHeal               = "heal"
	VerbHealHero           = "heal_hero"
	VerbHealVillain        = "heal_villain"
	VerbStun               = "stun"
	VerbStunChosen         = "stun_chosen"
	VerbConfuse            = "confuse"
	VerbGiveTough          = "give_tough"
	VerbExhaust            = "exhaust"
	VerbReady              = "ready"
	VerbFlipAlterEgo       = "flip_alter_ego"
	VerbDiscardRandom      = "discard_random"
	VerbDiscardRandomPlace = "discard_random_place_threat"
	VerbSurge              = "surge"
	VerbScheme             = "scheme"
	VerbVillainAttack      = "villain_attack"
	VerbVillainAndMinions  = "villain_and_minions_attack"
	VerbRemoveFromGame     = "remove_from_game"
	VerbReveal             = "reveal"
	VerbShuffleEncounter   = "shuffle_encounter"
	VerbDetach             = "detach"
	VerbToEncounterDiscard = "to_encounter_discard"
	VerbCancel             = "cancel"
	VerbPreventDamage      = "prevent_damage"
	VerbReduceAmount       = "reduce_amount"
	VerbTakeThreatAsDamage = "take_threat_as_damage"
	VerbAbsorbDamage       = "absorb_damage"
	VerbCounter            = "counter"
	VerbCostReduction      = "cost_reduction"
	VerbMillKeep           = "mill_keep"
	VerbFindAndReveal      = "find_and_reveal"
	VerbTakeRandomCard     = "take_random_card"
	VerbReturnAttached     = "return_attached"
	VerbAssignDamage       = "assign_damage"
	VerbNemesis            = "nemesis"
	VerbChoose             = "choose"
	VerbDiscardChosen      = "discard_chosen"
)

// Effect predicates (the values of Effect.When and Ability.When).
const (
	PredHero             = "hero"
	PredAlterEgo         = "alter_ego"
	PredHeroDamaged      = "hero_damaged"
	PredHeroConfused     = "hero_confused"
	PredVillainTough     = "villain_tough"
	PredSchemes          = "schemes"
	PredMinions          = "minions"
	PredNotExhausted     = "not_exhausted"
	PredExhausted        = "exhausted"
	PredCountersPositive = "counters_positive"
	PredHeroExhausted    = "hero_exhausted"
	PredHeroReady        = "hero_ready"
	PredSourceSelf       = "source_self"
	PredHostIsSource     = "host_is_source"
	PredHostIsTarget     = "host_is_target"
	PredBombScare        = "bomb_scare"
	PredNotBombScare     = "not_bomb_scare"
	PredVulture          = "vulture"
	PredUpgradesSupports = "upgrades_supports"
	PredAmountPositive   = "amount_positive"
	PredPaidPrefix       = "paid:"
	PredNotPaidPrefix    = "not_paid:"
)

// Effect selectors (the values of Ability.Target and Effect.Target).
const (
	SelEnemy              = "enemy"
	SelAllEnemy           = "all_enemy"
	SelMinion             = "minion"
	SelScheme             = "scheme"
	SelUpgradeSupport     = "upgrade_support"
	SelEnemyWithoutWebbed = "enemy_without_webbed_up"
	SelDamagedCharacter   = "damaged_character"
	SelEventSource        = "event_source"
	SelEventTarget        = "event_target"
	SelChosen             = "chosen"
	SelHero               = "hero"
	SelVillain            = "villain"
	SelSelf               = "self"
)

// Effect result markers.
const (
	ResultStun        = "stun"
	ResultEventAmount = "event_amount"
)

// Stat bonus expressions.
const BonusSideSchemes = "side_schemes"

// Well-known card codes.
const (
	CodeBombScare       = "01109"
	CodeBreakinAndTakin = "01107"
	CodeVulture         = "01167"
	CodeWebbedUp        = "01009"
)
