package solitaire

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/haruki7049/trump-center/internal/card"
)

func TestStateKey_RoundTrip(t *testing.T) {
	deck := card.NewDeck()
	card.Shuffle(deck, rand.New(rand.NewSource(1)))
	b, err := Deal(deck)
	if err != nil {
		t.Fatal(err)
	}
	// Move a few cards so every section (stock, waste, a pile with
	// several face-up cards, an empty pile) is exercised.
	b.DrawFromStock()
	b.DrawFromStock()
	b.Tableau[1].FaceUp = 2
	b.Tableau[0] = Pile{}

	got := decodeStateKey(b.stateKey(), b.Foundation)
	if !reflect.DeepEqual(got.Tableau, b.Tableau) || !reflect.DeepEqual(got.Stock, b.Stock) || !reflect.DeepEqual(got.Waste, b.Waste) {
		t.Errorf("decodeStateKey(stateKey()) = %+v; want %+v", got, b)
	}
}
