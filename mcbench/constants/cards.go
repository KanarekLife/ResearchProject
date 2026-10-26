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
	// TypeEnvironment is an encounter card that stays in play and changes
	// the rules (Ultron Drones).
	TypeEnvironment = "environment"
)

// AttachEnvironment is an attachment's attach_to value for the environment.
const AttachEnvironment = "environment"

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
	TrigAttacked              = "attacked"
	TrigAttackDamaged         = "attack_damaged"
	TrigMinionEnters          = "minion_enters"
	TrigEntersPlay            = "enters_play"
	TrigEnemyDamaged          = "enemy_damaged"
	TrigVillainSchemed        = "villain_schemed"
	TrigStepOneThreat         = "step_one_threat"
	TrigHeroWouldTakeDamage   = "hero_would_take_damage"
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
	AbilitySetup     = "setup"
	AbilityResponse  = "response"
	// AbilityBoostFlip is a "Boost" ability that resolves as the boost card
	// is flipped; AbilityBoost runs for each character the activation damages.
	AbilityBoostFlip = "boost_flip"
)

// Stats a stat ability may modify. The hero_* stats are bonuses a card you
// control gives your hero.
const (
	StatATK      = "atk"
	StatTHW      = "thw"
	StatHandSize = "hand_size"
	StatHeroATK  = "hero_atk"
	StatHeroTHW  = "hero_thw"
	StatHeroDEF  = "hero_def"
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
	VerbThwartEach         = "thwart_each"
	VerbAttackMill         = "attack_mill"
	VerbDamageHeroMill     = "damage_hero_mill"
	VerbDamageAlly         = "damage_ally"
	VerbDamageAllEnemies   = "damage_all_enemies"
	VerbDamageAllChars     = "damage_all_characters"
	VerbDamageFriendly     = "damage_friendly"
	VerbTempBoost          = "temp_boost"
	VerbGainTrait          = "gain_trait"
	VerbExhaustUpgrades    = "exhaust_upgrades"
	VerbExhaustAllies      = "exhaust_allies"
	VerbExhaustCharacters  = "exhaust_characters"
	VerbMill               = "mill"
	VerbDrone              = "drone"
	VerbLookKeep           = "look_keep"
	VerbReturnFromDiscard  = "return_from_discard"
	VerbDiscardHandChoice  = "discard_hand_choice"
	VerbDiscardHandRes     = "discard_hand_resource"
	VerbPutIntoPlay        = "put_into_play"
	VerbDiscardUntilMinion = "discard_until_minion"
	VerbSearchMinion       = "search_minion"
	VerbEngageSelf         = "engage_self"
	VerbMinionsAttack      = "minions_attack"
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
	PredAerial           = "aerial"
	PredNotAerial        = "not_aerial"
	PredDefeated         = "defeated"
	PredSourceHero       = "source_hero"
	PredTargetMinion     = "target_minion"
	PredTargetSelf       = "target_self"
	PredTargetHero       = "target_hero"
	PredUndefended       = "undefended"
	PredHandNotEmpty     = "hand_not_empty"
	PredDrones           = "drones"
	PredTargetInPlay     = "target_in_play"
	PredUpgrades         = "upgrades"
	PredMilledPrefix     = "milled:"
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
	SelAlly               = "ally"
	SelAllyWithoutThis    = "ally_without_this"
	SelDiscardAlly        = "discard_ally"
	SelUpgrade            = "upgrade"
	SelSideSchemes        = "side_schemes"
	SelFriendly           = "friendly"
)

// Effect result markers.
const (
	ResultStun        = "stun"
	ResultEventAmount = "event_amount"
	// ResultThreat places 1 threat on the main scheme if the attack damages.
	ResultThreat = "threat"
	// ResultMill discards the top card of your deck per damage dealt or
	// threat placed.
	ResultMill = "mill"
)

// Count expressions, for stat bonuses and an effect's per.
const (
	CountSideSchemes  = "side_schemes"
	CountTechUpgrades = "tech_upgrades"
	CountAerial       = "aerial"
	CountDrones       = "drones"
	CountUpgrades     = "upgrades"
	CountCounters     = "counters"
)

// Traits the rules refer to.
const (
	TraitAerial = "Aerial"
	TraitTech   = "Tech"
	TraitDrone  = "Drone"
)

// Well-known card codes.
const (
	// CodeDrone is the face a facedown Drone minion shows (not a printed card).
	CodeDrone           = "drone"
	CodeBombScare       = "01109"
	CodeBreakinAndTakin = "01107"
	CodeVulture         = "01167"
	CodeWebbedUp        = "01009"
)
