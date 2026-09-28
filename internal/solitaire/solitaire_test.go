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

func TestBoard_MoveWasteToFoundation(t *testing.T) {
	b := solitaire.Board{Waste: []card.Card{{Suit: card.Spade, Rank: card.Ace}}}

	if !b.MoveWasteToFoundation() {
		t.Fatal("expected the Ace to be movable to an empty foundation")
	}
	if len(b.Waste) != 0 {
		t.Errorf("expected waste to be empty, got %d cards", len(b.Waste))
	}
	if len(b.Foundation[card.Spade]) != 1 {
		t.Errorf("expected 1 card on the spade foundation, got %d", len(b.Foundation[card.Spade]))
	}

	if b.MoveWasteToFoundation() {
		t.Error("expected the move to fail on an empty waste")
	}
}

func TestBoard_MoveWasteToTableau(t *testing.T) {
	b := solitaire.Board{
		Waste: []card.Card{{Suit: card.Heart, Rank: card.Queen}},
		Tableau: [solitaire.TableauPileCount]solitaire.Pile{
			0: {Cards: []card.Card{{Suit: card.Spade, Rank: card.King}}, FaceUp: 1},
		},
	}

	if !b.MoveWasteToTableau(0) {
		t.Fatal("expected Queen of Hearts to be movable onto the King of Spades")
	}
	if len(b.Waste) != 0 {
		t.Errorf("expected waste to be empty, got %d cards", len(b.Waste))
	}
	if got := b.Tableau[0]; len(got.Cards) != 2 || got.FaceUp != 2 {
		t.Errorf("tableau pile 0 = %+v; want 2 cards, FaceUp=2", got)
	}
}

func TestBoard_MoveTableauToFoundation(t *testing.T) {
	b := solitaire.Board{
		Tableau: [solitaire.TableauPileCount]solitaire.Pile{
			0: {Cards: []card.Card{
				{Suit: card.Club, Rank: card.King},
				{Suit: card.Club, Rank: card.Ace},
			}, FaceUp: 1},
		},
	}

	if !b.MoveTableauToFoundation(0) {
		t.Fatal("expected the Ace to be movable to an empty foundation")
	}
	if got := b.Tableau[0]; len(got.Cards) != 1 || got.FaceUp != 1 {
		t.Errorf("tableau pile 0 = %+v; want 1 card, FaceUp=1 (previous card flipped up)", got)
	}

	// The King is face-down, so it must not be movable.
	if b.MoveTableauToFoundation(0) {
		t.Error("expected the face-down King to not be movable")
	}
}

func TestBoard_MoveTableauToTableauRun_SingleCard(t *testing.T) {
	b := solitaire.Board{
		Tableau: [solitaire.TableauPileCount]solitaire.Pile{
			0: {Cards: []card.Card{
				{Suit: card.Club, Rank: card.Queen},
				{Suit: card.Heart, Rank: card.Ten},
			}, FaceUp: 1},
			1: {Cards: []card.Card{{Suit: card.Spade, Rank: card.Jack}}, FaceUp: 1},
		},
	}

	// cardIndex 1 (the pile's last index) is a run of just the top card.
	if !b.MoveTableauToTableauRun(0, 1, 1) {
		t.Fatal("expected the Ten of Hearts to be movable onto the Jack of Spades")
	}
	if got := b.Tableau[0]; len(got.Cards) != 1 || got.FaceUp != 1 {
		t.Errorf("source pile = %+v; want 1 card, FaceUp=1", got)
	}
	if got := b.Tableau[1]; len(got.Cards) != 2 {
		t.Errorf("destination pile = %+v; want 2 cards", got)
	}
}

func TestBoard_MoveTableauToTableauRun(t *testing.T) {
	b := solitaire.Board{
		Tableau: [solitaire.TableauPileCount]solitaire.Pile{
			0: {Cards: []card.Card{
				{Suit: card.Diamond, Rank: card.King}, // face-down
				{Suit: card.Club, Rank: card.Eight},
				{Suit: card.Heart, Rank: card.Seven},
				{Suit: card.Club, Rank: card.Six},
			}, FaceUp: 3},
			1: {Cards: []card.Card{{Suit: card.Heart, Rank: card.Nine}}, FaceUp: 1},
		},
	}

	// cardIndex 1 is the Eight of Clubs, the bottom of a valid 8-7-6 run.
	if !b.MoveTableauToTableauRun(0, 1, 1) {
		t.Fatal("expected the 8-7-6 run to be movable onto the 9 of Hearts")
	}
	if got := b.Tableau[0]; len(got.Cards) != 1 || got.FaceUp != 1 {
		t.Errorf("source pile = %+v; want 1 card, FaceUp=1 (the King flipped face-up)", got)
	}
	if got := b.Tableau[1]; len(got.Cards) != 4 || got.FaceUp != 4 {
		t.Errorf("destination pile = %+v; want 4 cards, FaceUp=4", got)
	}
}

func TestBoard_MoveTableauToTableauRun_InvalidRun(t *testing.T) {
	b := solitaire.Board{
		Tableau: [solitaire.TableauPileCount]solitaire.Pile{
			// Eight then Six is not a valid run (skips the Seven).
			0: {Cards: []card.Card{
				{Suit: card.Club, Rank: card.Eight},
				{Suit: card.Club, Rank: card.Six},
			}, FaceUp: 2},
			1: {Cards: []card.Card{{Suit: card.Spade, Rank: card.Nine}}, FaceUp: 1},
		},
	}

	if b.MoveTableauToTableauRun(0, 0, 1) {
		t.Error("expected an invalid run to not be movable")
	}
}

func TestBoard_MoveTableauToTableauRun_FaceDownCardIndex(t *testing.T) {
	b := solitaire.Board{
		Tableau: [solitaire.TableauPileCount]solitaire.Pile{
			0: {Cards: []card.Card{
				{Suit: card.Club, Rank: card.Eight}, // face-down
				{Suit: card.Heart, Rank: card.Seven},
			}, FaceUp: 1},
			1: {Cards: []card.Card{{Suit: card.Spade, Rank: card.Nine}}, FaceUp: 1},
		},
	}

	if b.MoveTableauToTableauRun(0, 0, 1) {
		t.Error("expected picking a run starting at a face-down card to fail")
	}
}

func TestBoard_MoveTableauToTableauRun_SamePile(t *testing.T) {
	b := solitaire.Board{
		Tableau: [solitaire.TableauPileCount]solitaire.Pile{
			0: {Cards: []card.Card{{Suit: card.Heart, Rank: card.Seven}}, FaceUp: 1},
		},
	}

	if b.MoveTableauToTableauRun(0, 0, 0) {
		t.Error("expected moving a run onto its own pile to fail")
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
