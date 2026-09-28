package solitaire_test

import (
	"testing"

	"github.com/haruki7049/trump-center/internal/card"
	"github.com/haruki7049/trump-center/internal/solitaire"
)

func TestPile_Top(t *testing.T) {
	empty := solitaire.Pile{}
	if _, ok := empty.Top(); ok {
		t.Errorf("expected ok=false for empty pile")
	}

	want := card.Card{Suit: card.Spade, Rank: card.Five}
	pile := solitaire.Pile{Cards: []card.Card{{Suit: card.Heart, Rank: card.Two}, want}}
	got, ok := pile.Top()
	if !ok || got != want {
		t.Errorf("Top() = (%+v, %v); want (%+v, true)", got, ok, want)
	}
}

func TestDeal(t *testing.T) {
	deck := card.NewDeck()

	b, err := solitaire.Deal(deck)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	total := len(b.Stock)
	for i, pile := range b.Tableau {
		wantSize := i + 1
		if len(pile.Cards) != wantSize {
			t.Errorf("tableau pile %d has %d cards; want %d", i, len(pile.Cards), wantSize)
		}
		if pile.FaceUp != 1 {
			t.Errorf("tableau pile %d has FaceUp=%d; want 1", i, pile.FaceUp)
		}
		total += len(pile.Cards)
	}

	if total != 52 {
		t.Errorf("total dealt cards = %d; want 52", total)
	}
}

func TestDeal_WrongDeckSize(t *testing.T) {
	_, err := solitaire.Deal(card.NewDeck()[:51])
	if err == nil {
		t.Fatal("expected an error for a deck that is not 52 cards")
	}
}

func TestBoard_DrawFromStock(t *testing.T) {
	b := solitaire.Board{
		Stock: []card.Card{
			{Suit: card.Spade, Rank: card.Ace},
			{Suit: card.Spade, Rank: card.Two},
			{Suit: card.Spade, Rank: card.Three},
		},
	}

	b.DrawFromStock()
	want := card.Card{Suit: card.Spade, Rank: card.Three}
	if got, ok := b.WasteTop(); !ok {
		t.Fatal("expected a top waste card after drawing")
	} else if got != want {
		t.Errorf("waste top = %+v; want %+v", got, want)
	}
	if len(b.Stock) != 2 {
		t.Errorf("stock len = %d; want 2", len(b.Stock))
	}

	b.DrawFromStock()
	b.DrawFromStock()
	if len(b.Stock) != 0 {
		t.Fatalf("expected stock to be empty, got %d cards", len(b.Stock))
	}
	if len(b.Waste) != 3 {
		t.Fatalf("expected 3 cards in waste, got %d", len(b.Waste))
	}

	// Stock is now empty: drawing should recycle the waste back into the
	// stock instead of drawing a card.
	b.DrawFromStock()
	if len(b.Waste) != 0 {
		t.Errorf("expected waste to be emptied by recycling, got %d cards", len(b.Waste))
	}
	if len(b.Stock) != 3 {
		t.Fatalf("expected 3 cards back in stock, got %d", len(b.Stock))
	}

	// Recycling must reproduce the original draw order: the first card
	// drawn before was Three, so it must be the first one drawn again.
	b.DrawFromStock()
	if got, _ := b.WasteTop(); got != (card.Card{Suit: card.Spade, Rank: card.Three}) {
		t.Errorf("first card drawn after recycling = %+v; want Three", got)
	}
}

func TestBoard_CanMoveToFoundation(t *testing.T) {
	var b solitaire.Board

	if !b.CanMoveToFoundation(card.Card{Suit: card.Spade, Rank: card.Ace}) {
		t.Errorf("expected Ace to be movable onto an empty foundation")
	}
	if b.CanMoveToFoundation(card.Card{Suit: card.Spade, Rank: card.Two}) {
		t.Errorf("expected non-Ace to be rejected on an empty foundation")
	}

	b.Foundation[card.Spade] = []card.Card{{Suit: card.Spade, Rank: card.Ace}}
	if !b.CanMoveToFoundation(card.Card{Suit: card.Spade, Rank: card.Two}) {
		t.Errorf("expected the next rank of the same suit to be movable")
	}
	if b.CanMoveToFoundation(card.Card{Suit: card.Spade, Rank: card.Three}) {
		t.Errorf("expected skipping a rank to be rejected")
	}
}

func TestBoard_CanMoveToTableau(t *testing.T) {
	var b solitaire.Board

	if !b.CanMoveToTableau(card.Card{Suit: card.Spade, Rank: card.King}, 0) {
		t.Errorf("expected King to be movable onto an empty tableau pile")
	}
	if b.CanMoveToTableau(card.Card{Suit: card.Spade, Rank: card.Queen}, 0) {
		t.Errorf("expected non-King to be rejected on an empty tableau pile")
	}

	b.Tableau[0] = solitaire.Pile{
		Cards:  []card.Card{{Suit: card.Spade, Rank: card.King}},
		FaceUp: 1,
	}
	if !b.CanMoveToTableau(card.Card{Suit: card.Heart, Rank: card.Queen}, 0) {
		t.Errorf("expected an alternating-color, descending-rank card to be movable")
	}
	if b.CanMoveToTableau(card.Card{Suit: card.Club, Rank: card.Queen}, 0) {
		t.Errorf("expected a same-color card to be rejected")
	}
	if b.CanMoveToTableau(card.Card{Suit: card.Heart, Rank: card.Jack}, 0) {
		t.Errorf("expected skipping a rank to be rejected")
	}
}
