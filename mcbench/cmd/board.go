package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"mcbench/game/cards"
	"mcbench/game/session"
)

// boardRenderScript is the POC board renderer (run from mcbench/): it draws
// the board GUI's Board component with headless Chromium.
const boardRenderScript = "boardgui/web/render/shot.mjs"

// boardPictures renders the board as a PNG for the vision variant.
type boardPictures struct {
	codes map[string][]string
}

func newBoardPictures() boardPictures {
	codes := map[string][]string{}
	for _, d := range cards.All() {
		codes[d.Name] = append(codes[d.Name], d.Code)
	}
	return boardPictures{codes: codes}
}

func (b boardPictures) Render(v session.View) (string, error) {
	f, err := os.CreateTemp("", "board-*.png")
	if err != nil {
		return "", err
	}
	f.Close()
	defer os.Remove(f.Name())
	payload, err := json.Marshal(map[string]any{"view": v, "codes": b.codes})
	if err != nil {
		return "", err
	}
	cmd := exec.Command("node", boardRenderScript, f.Name())
	cmd.Stdin = bytes.NewReader(payload)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("render board: %w: %s", err, out)
	}
	png, err := os.ReadFile(f.Name())
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}
