package cards

import (
	e "mcbench/game/engine"
)

func cardType(s string) e.CardType {
	t, ok := cardTypes[s]
	if !ok {
		panic("cards: unknown card type " + s)
	}
	return t
}

func resource(s string) e.Resource {
	r, ok := resourceNames[s]
	if !ok {
		panic("cards: unknown resource " + s)
	}
	return r
}

func resources(ss []string) []e.Resource {
	out := make([]e.Resource, len(ss))
	for i, s := range ss {
		out[i] = resource(s)
	}
	return out
}

func trigger(s string) e.Trigger {
	t, ok := triggers[s]
	if !ok {
		panic("cards: unknown trigger " + s)
	}
	return t
}
