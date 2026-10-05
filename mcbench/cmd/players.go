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
func buildPlayer(name string, client inference.Client, label string, maxSteps int, instr instruction.Instructions) (player.Player, error) {
	switch name {
	case playerModel:
		if client == nil {
			return nil, fmt.Errorf("-player model needs a model id")
		}
		return &model.Model{Client: client, Label: label, MaxSteps: maxSteps, Instructions: instr}, nil
	case playerHeuristic:
		return heuristic.Heuristic{}, nil
	case playerRandom:
		return heuristic.Random{Seed: 1}, nil
	case playerFirst:
		return heuristic.First{}, nil
	}
	return nil, fmt.Errorf("unknown player %q", name)
}
