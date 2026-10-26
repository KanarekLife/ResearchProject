package model

import "mcbench/game/session"

// stateView is the board as get_state returns it to the model: the session
// View without the later villain stages (get_card has them), with the discard
// piles as counts, and with empty lists and zero counts left out. Like a
// decision, every get_state result stays in the conversation and is resent on
// every later request.
type stateView struct {
	Round                int                   `json:"round"`
	Phase                string                `json:"phase"`
	Hero                 session.Hero          `json:"hero"`
	Hand                 []session.Card        `json:"hand,omitempty"`
	InPlay               []session.Card        `json:"in_play,omitempty"`
	DeckSize             int                   `json:"deck_size,omitempty"`
	DiscardSize          int                   `json:"discard_size,omitempty"`
	FaceDownEncounters   int                   `json:"face_down_encounter_cards,omitempty"`
	Villain              session.Enemy         `json:"villain"`
	MainScheme           session.Scheme        `json:"main_scheme"`
	SideSchemes          []session.Scheme      `json:"side_schemes,omitempty"`
	Minions              []session.Enemy       `json:"minions,omitempty"`
	Environment          []session.Environment `json:"environment,omitempty"`
	EncounterDeckSize    int                   `json:"encounter_deck_size,omitempty"`
	EncounterDiscardSize int                   `json:"encounter_discard_size,omitempty"`
	AccelerationTokens   int                   `json:"acceleration_tokens,omitempty"`
	NextCardCostReduced  int                   `json:"next_card_cost_reduction,omitempty"`
}

func compactState(v session.View) stateView {
	return stateView{
		Round: v.Round, Phase: v.Phase, Hero: v.Hero, Hand: v.Hand, InPlay: v.InPlay,
		DeckSize: v.DeckSize, DiscardSize: len(v.Discard), FaceDownEncounters: v.FaceDownEncounters,
		Villain: v.Villain, MainScheme: v.MainScheme, SideSchemes: v.SideSchemes, Minions: v.Minions, Environment: v.Environment,
		EncounterDeckSize: v.EncounterDeckSize, EncounterDiscardSize: len(v.EncounterDiscard),
		AccelerationTokens: v.AccelerationTokens, NextCardCostReduced: v.NextCardCostReduced,
	}
}

// BoardState is the board as get_state shows it, for players that are not
// driven through the tools.
func BoardState(s *session.Session) any { return compactState(s.View()) }
