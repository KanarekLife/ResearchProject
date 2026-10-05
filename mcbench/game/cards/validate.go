package cards

import "fmt"

// validate checks a card document's predicates so a typo fails when the
// cards load, not in the middle of a game.
func validate(d *CardDoc) error {
	for _, a := range d.Abilities {
		for _, list := range [][]string{a.When, a.Usable} {
			if err := checkPredicates(list); err != nil {
				return fmt.Errorf("%s (%s): %w", d.Name, d.Code, err)
			}
		}
		if err := checkEffects(a.Effects); err != nil {
			return fmt.Errorf("%s (%s): %w", d.Name, d.Code, err)
		}
	}
	if d.Back != nil {
		return validate(d.Back)
	}
	return nil
}

func checkPredicates(preds []string) error {
	for _, p := range preds {
		if predicate(p) == nil {
			return fmt.Errorf("unknown predicate %q", p)
		}
	}
	return nil
}

func checkEffects(effects []Effect) error {
	for _, eff := range effects {
		if err := checkPredicates(eff.When); err != nil {
			return err
		}
		for _, branch := range [][]Effect{eff.IfZero, eff.IfAlready, eff.IfEmpty} {
			if err := checkEffects(branch); err != nil {
				return err
			}
		}
		for _, o := range eff.Options {
			if err := checkPredicates(o.When); err != nil {
				return err
			}
			if err := checkEffects(o.Effects); err != nil {
				return err
			}
		}
	}
	return nil
}
