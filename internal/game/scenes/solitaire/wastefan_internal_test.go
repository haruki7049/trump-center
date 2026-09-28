package solitaire

import (
	"testing"

	"github.com/haruki7049/trump-center/internal/card"
	"github.com/haruki7049/trump-center/internal/solitaire"
)

func TestSolitaireScene_SyncComponents_FansWasteCards(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{
		Waste: []card.Card{
			{Suit: card.Spade, Rank: card.Ace},
			{Suit: card.Spade, Rank: card.Two},
			{Suit: card.Spade, Rank: card.Three},
			{Suit: card.Spade, Rank: card.Four},
			{Suit: card.Spade, Rank: card.Five},
		},
	})
	s.syncComponents()

	for i := range wasteFanCount {
		b, ok := s.wastePile.CardBoundsAt(i)
		if !ok {
			t.Fatalf("expected card %d to be shown in the fan", i)
		}
		if want := wasteOriginX + i*wasteFanOffsetX; b.Min.X != want {
			t.Errorf("card %d X = %d; want %d", i, b.Min.X, want)
		}
	}
	if _, ok := s.wastePile.CardBoundsAt(wasteFanCount); ok {
		t.Errorf("expected only %d cards to be shown even though the waste has more", wasteFanCount)
	}

	top, ok := s.wastePile.TopBounds()
	if !ok {
		t.Fatal("expected a top card")
	}
	if want := wasteOriginX + (wasteFanCount-1)*wasteFanOffsetX; top.Min.X != want {
		t.Errorf("TopBounds X = %d; want %d (the most recently drawn card)", top.Min.X, want)
	}
}

func TestSolitaireScene_SyncComponents_WasteFanWithFewerCardsThanCount(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{
		Waste: []card.Card{{Suit: card.Club, Rank: card.King}},
	})
	s.syncComponents()

	top, ok := s.wastePile.TopBounds()
	if !ok {
		t.Fatal("expected a top card")
	}
	if top.Min.X != wasteOriginX {
		t.Errorf("with a single waste card, X = %d; want %d", top.Min.X, wasteOriginX)
	}
	if _, ok := s.wastePile.CardBoundsAt(1); ok {
		t.Error("expected only one card to be shown")
	}
}

func TestSolitaireScene_SyncComponents_WasteFanRevealsCardBehindDrag(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{
		Waste: []card.Card{
			{Suit: card.Heart, Rank: card.Nine},
			{Suit: card.Heart, Rank: card.Ten},
		},
	})
	s.syncComponents()

	top, ok := s.wastePile.TopBounds()
	if !ok {
		t.Fatal("expected a top card")
	}
	s.mediator.Handle(Event{Type: EventPointerDown, X: top.Min.X, Y: top.Min.Y})
	if _, ok := s.mediator.Dragging(); !ok {
		t.Fatal("expected a drag to start")
	}

	s.syncComponents()

	if _, ok := s.wastePile.CardBoundsAt(1); ok {
		t.Error("expected the dragged card to not be drawn in the fan")
	}
	if _, ok := s.wastePile.CardBoundsAt(0); !ok {
		t.Error("expected the card behind the dragged one to still be shown")
	}
}
