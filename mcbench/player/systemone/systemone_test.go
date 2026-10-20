package systemone

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"mcbench/constants"
	"mcbench/data"
	"mcbench/game/scenario"
	"mcbench/game/session"
	"mcbench/integrations/systemone"
)

// A player that always takes the first option still pays for its cards in the
// same step and never backs out of a payment.
func TestPlayNeverCancelsAPayment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"answers":{"decision":{"choice":"1","probabilities":{"1":1}}},"usage":{}}`))
	}))
	defer srv.Close()
	scs, err := scenario.LoadAll(data.FS)
	if err != nil {
		t.Fatal(err)
	}
	sc := scs[0]
	s, err := session.New(sc, sc.Seeds[0], session.Options{MaxRounds: sc.MaxRounds, MaxDecisions: sc.MaxDecisions})
	if err != nil {
		t.Fatal(err)
	}
	p := &SystemOne{Client: systemone.New(systemone.Config{BaseURL: srv.URL}), RecentEvents: 5}
	if _, err := p.Play(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if s.Status() == constants.AwaitingDecision {
		t.Fatal("game did not finish")
	}
	paid := 0
	for _, c := range s.Choices {
		if c.Key == constants.KeyCancel {
			t.Fatalf("cancelled at decision %d", len(s.Choices))
		}
		if c.Kind == constants.KindPayment {
			paid++
		}
	}
	if paid == 0 {
		t.Error("the game never paid for a card")
	}
}
