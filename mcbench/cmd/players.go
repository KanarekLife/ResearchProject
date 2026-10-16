package main

import (
	"fmt"

	"mcbench/integrations/inference"
	"mcbench/player"
	"mcbench/player/heuristic"
	"mcbench/player/instruction"
	"mcbench/player/model"
)

// buildPlayer builds the named player. Model settings are configured here,
// in cmd, and passed to the model player; the scripted players ignore them.
func buildPlayer(o runSettings, client inference.Client, instr instruction.Instructions) (player.Player, error) {
	switch o.player {
	case playerModel:
		if client == nil {
			return nil, fmt.Errorf("-player model needs a model id")
		}
		return &model.Model{Client: client, Label: modelLabel(o), MaxFailures: o.maxFailures, Instructions: instr}, nil
	case playerHeuristic:
		return heuristic.Heuristic{}, nil
	case playerRandom:
		return heuristic.Random{Seed: randomPlayerSeed}, nil
	case playerFirst:
		return heuristic.First{}, nil
	}
	return nil, fmt.Errorf("unknown player %q", o.player)
}

// modelLabel names the model in results. The reasoning setting is part of the
// name so runs at different settings are reported as separate players.
func modelLabel(o runSettings) string {
	label := o.modelID
	if o.effort != "" {
		label += labelEffort + o.effort
	}
	if o.budget > 0 {
		label += fmt.Sprintf(labelBudget+"%d", o.budget)
	}
	return label
}

// Suffixes modelLabel adds for the reasoning setting.
const (
	labelEffort = "@effort-"
	labelBudget = "@budget-"
)
