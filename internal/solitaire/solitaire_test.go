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
