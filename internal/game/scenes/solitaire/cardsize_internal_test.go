package solitaire

import (
	"testing"

	"github.com/haruki7049/trump-center/internal/card"
	"github.com/haruki7049/trump-center/internal/solitaire"
)

func TestTableauCardY(t *testing.T) {
	faceUpTop := tableauOriginY + 2*faceDownOffsetY

	tests := []struct {
		name                          string
		faceDownCount, faceUpCount, i int
		want                          int
	}{
		{"face-down cards step tightly", 2, 3, 1, tableauOriginY + faceDownOffsetY},
		{"first face-up card follows the face-down ones", 2, 3, 2, faceUpTop},
		{"short pile uses the full face-up step", 2, 3, 4, faceUpTop + 2*faceUpOffsetY},
		{"single face-up card", 0, 1, 0, tableauOriginY},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tableauCardY(tt.faceDownCount, tt.faceUpCount, tt.i); got != tt.want {
				t.Errorf("tableauCardY(%d, %d, %d) = %d; want %d", tt.faceDownCount, tt.faceUpCount, tt.i, got, tt.want)
			}
		})
	}
}

// TestTableauCardY_TallestPileFits checks the tallest pile Klondike can
// produce — six face-down cards under a full King-to-Ace run — still ends
// above tableauMaxBottom, by squeezing its face-up step.
func TestTableauCardY_TallestPileFits(t *testing.T) {
	const faceDown, faceUp = 6, 13
	last := faceDown + faceUp - 1

	if bottom := tableauCardY(faceDown, faceUp, last) + cardHeight; bottom > tableauMaxBottom {
		t.Errorf("tallest pile's bottom = %d; want <= %d", bottom, tableauMaxBottom)
	}

	step := tableauCardY(faceDown, faceUp, last) - tableauCardY(faceDown, faceUp, last-1)
	if step >= faceUpOffsetY || step < minFaceUpOffsetY {
		t.Errorf("squeezed face-up step = %d; want in [%d, %d)", step, minFaceUpOffsetY, faceUpOffsetY)
	}
}

func TestSolitaireScene_SyncComponents_UsesTableauCardY(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{})
	s.board.Tableau[3] = solitaire.Pile{Cards: []card.Card{
		{Suit: card.Spade, Rank: card.Two},  // face-down
		{Suit: card.Heart, Rank: card.Nine}, // face-down
		{Suit: card.Club, Rank: card.Eight},
		{Suit: card.Diamond, Rank: card.Seven},
	}, FaceUp: 2}
	s.syncComponents()

	for i := range 4 {
		b, ok := s.tableauPiles[3].CardBoundsAt(i)
		if !ok {
			t.Fatalf("expected card %d to be drawn", i)
		}
		if want := tableauCardY(2, 2, i); b.Min.Y != want {
			t.Errorf("card %d Y = %d; want %d", i, b.Min.Y, want)
		}
	}
}
