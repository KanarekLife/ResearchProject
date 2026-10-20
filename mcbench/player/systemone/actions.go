package systemone

import (
	"fmt"

	"mcbench/constants"
	"mcbench/game/session"
)

// action is one thing the classifier can pick: an option of the decision, or
// an option together with how to pay for it.
type action struct {
	option  session.Option
	payment *session.Option // the way to pay, when the option leads to a payment decision
	text    string
}

// actions lists what can be picked at d. A card with several ways to pay is
// one action per way, naming the cards discarded, so the classifier chooses the
// payment with the card and can neither back out of it nor loop on doing so.
func actions(s *session.Session, d *session.Decision) []action {
	var out []action
	for _, o := range d.Options {
		if o.Key == constants.KeyCancel {
			continue
		}
		next, err := s.Preview(o.ID)
		if err != nil || next == nil || next.Kind != constants.KindPayment {
			out = append(out, action{option: o, text: o.Text})
			continue
		}
		for _, pay := range next.Options {
			if pay.Key != constants.KeyCancel {
				out = append(out, action{option: o, payment: &pay, text: fmt.Sprintf("%s, paying with %s", o.Text, pay.Text)})
			}
		}
	}
	if len(out) == 0 { // only a way back was offered
		for _, o := range d.Options {
			out = append(out, action{option: o, text: o.Text})
		}
	}
	return out
}

// play makes the action in the game.
func (a action) play(s *session.Session, reasoning string) error {
	if _, err := s.Choose(a.option.ID, reasoning); err != nil {
		return err
	}
	if a.payment == nil {
		return nil
	}
	d := s.Decision()
	if d == nil {
		return fmt.Errorf("the game ended before paying for %q", a.option.Text)
	}
	for _, o := range d.Options {
		if o.Key == a.payment.Key {
			_, err := s.Choose(o.ID, reasoning)
			return err
		}
	}
	return fmt.Errorf("payment %q is not offered", a.payment.Key)
}
