// Package cards holds the card definitions implemented so far. Only cards a
// benchmark scenario needs are implemented; see AGENTS.md.
package cards

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	e "mcbench/engine"
)

var (
	byCode = map[string]*e.CardDef{}
	byName = map[string][]*e.CardDef{}
)

func add(d *e.CardDef) {
	if d.IsPlayerCard() && d.Aspect == "" {
		d.Aspect = aspectOf(d.Code)
	}
	if _, dup := byCode[d.Code]; dup {
		panic("cards: duplicate code " + d.Code)
	}
	byCode[d.Code] = d
	byName[strings.ToLower(d.Name)] = append(byName[strings.ToLower(d.Name)], d)
	if d.Back != nil {
		byCode[d.Back.Code] = d.Back
	}
}

// Get resolves a card by code ("01005") or by unique name ("Swinging Web Kick").
func Get(ref string) (*e.CardDef, error) {
	if d, ok := byCode[ref]; ok {
		return d, nil
	}
	ds := byName[strings.ToLower(ref)]
	switch len(ds) {
	case 0:
		return nil, fmt.Errorf("unknown card %q (not implemented?)", ref)
	case 1:
		return ds[0], nil
	}
	codes := make([]string, len(ds))
	for i, d := range ds {
		codes[i] = d.Code
	}
	return nil, fmt.Errorf("card name %q is ambiguous; use a code: %s", ref, strings.Join(codes, ", "))
}

// All returns every implemented card, sorted by code.
func All() []*e.CardDef {
	out := make([]*e.CardDef, 0, len(byCode))
	for _, d := range byCode {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

func rs(r ...e.Resource) []e.Resource { return r }

func itoa(n int) string { return strconv.Itoa(n) }
