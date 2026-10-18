package main

import (
	"fmt"
	"os/exec"

	"mcbench/integrations/inference"
	"mcbench/player"
	"mcbench/player/heuristic"
	"mcbench/player/instruction"
	"mcbench/player/model"
	"mcbench/player/retrieval"
)

// buildPlayer builds the named player. Model settings are configured here,
// in cmd, and passed to the model player; the scripted players ignore them.
func buildPlayer(o runSettings, client inference.Client, instr instruction.Instructions) (player.Player, error) {
	switch o.player {
	case playerModel:
		if client == nil {
			return nil, fmt.Errorf("-player model needs a model id")
		}
		if o.vision && o.api != apiOpenAI {
			return nil, fmt.Errorf("-vision needs -api %s", apiOpenAI)
		}
		m := &model.Model{Client: client, Label: modelLabel(o), MaxFailures: o.maxFailures, Instructions: instr}
		if o.rag {
			text, err := ragSourceText()
			if err != nil {
				return nil, err
			}
			m.Retriever = retrieval.New(text, ragChunkWords)
			m.Instructions = instruction.Instructions{Names: instr.Names, SHA256: instr.SHA256}
		}
		if o.vision {
			m.Board = newBoardPictures()
		}
		return m, nil
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
	if o.rag {
		label += labelRAG
	}
	if o.vision {
		label += labelVision
	}
	return label
}

// ragChunkWords is the size of the rule chunks the RAG variant retrieves from.
const ragChunkWords = 120

// ragSourcePDF is the rules reference the RAG variant retrieves from. Relative
// to mcbench/, where the commands run.
const ragSourcePDF = "../docs/references/rules_reference_v18.pdf"

// ragSourceText is the official rules reference as plain text, the source the
// RAG variant retrieves from. pdftotext (poppler-utils) reads the PDF.
func ragSourceText() (string, error) {
	out, err := exec.Command("pdftotext", "-enc", "UTF-8", ragSourcePDF, "-").Output()
	if err != nil {
		return "", fmt.Errorf("rules reference %s (needs pdftotext): %w", ragSourcePDF, err)
	}
	return string(out), nil
}

// Suffixes modelLabel adds for the reasoning setting.
const (
	labelEffort = "@effort-"
	labelBudget = "@budget-"
	labelRAG    = "@rag"
	labelVision = "@vision"
)
