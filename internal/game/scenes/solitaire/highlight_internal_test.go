package solitaire

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/haruki7049/trump-center/internal/card"
	"github.com/haruki7049/trump-center/internal/solitaire"
)

func TestMediator_ValidDropTargets_NotDragging(t *testing.T) {
	s := newTestScene(t)
	if _, ok := s.mediator.ValidDropTargets(); ok {
		t.Error("expected no drop targets while not dragging")
	}
}

func TestMediator_ValidDropTargets_FromWaste(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{
		Waste: []card.Card{{Suit: card.Heart, Rank: card.Ace}},
	})
	// A red Ace can go onto its (empty) foundation, but onto no tableau
	// pile: empty piles only take Kings, and pile 0's top is a same-color
	// Two.
	s.board.Tableau[0] = solitaire.Pile{Cards: []card.Card{{Suit: card.Diamond, Rank: card.Two}}, FaceUp: 1}
	s.syncComponents()

	top, _ := s.wastePile.TopBounds()
	s.mediator.Handle(Event{Type: EventPointerDown, X: top.Min.X, Y: top.Min.Y})

	got, ok := s.mediator.ValidDropTargets()
	if !ok {
		t.Fatal("expected drop targets while dragging")
	}
	var want DropTargets
	want.Foundation[card.Heart] = true
	if got != want {
		t.Errorf("ValidDropTargets() = %+v; want %+v", got, want)
	}
}

func TestMediator_ValidDropTargets_FromTableau(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{})
	s.board.Tableau[0] = solitaire.Pile{Cards: []card.Card{{Suit: card.Heart, Rank: card.Nine}}, FaceUp: 1}
	s.board.Tableau[1] = solitaire.Pile{Cards: []card.Card{
		{Suit: card.Club, Rank: card.Eight},
		{Suit: card.Diamond, Rank: card.Seven},
	}, FaceUp: 2}
	s.board.Tableau[2] = solitaire.Pile{Cards: []card.Card{{Suit: card.Diamond, Rank: card.Nine}}, FaceUp: 1}
	s.board.Tableau[3] = solitaire.Pile{Cards: []card.Card{{Suit: card.Spade, Rank: card.Nine}}, FaceUp: 1}
	s.board.Foundation[card.Diamond] = []card.Card{
		{Suit: card.Diamond, Rank: card.Ace}, {Suit: card.Diamond, Rank: card.Two},
		{Suit: card.Diamond, Rank: card.Three}, {Suit: card.Diamond, Rank: card.Four},
		{Suit: card.Diamond, Rank: card.Five}, {Suit: card.Diamond, Rank: card.Six},
	}
	s.syncComponents()

	tests := []struct {
		name      string
		cardIndex int
		want      DropTargets
	}{
		{
			// The 8-7 run fits on either red Nine, but never the
			// foundation (it's more than one card) or the black Nine.
			name:      "run",
			cardIndex: 0,
			want:      DropTargets{Tableau: [solitaire.TableauPileCount]bool{0: true, 2: true}},
		},
		{
			// The Seven of Diamonds alone fits on its foundation (after
			// the Six), but on no Nine.
			name:      "single card",
			cardIndex: 1,
			want:      DropTargets{Foundation: [4]bool{card.Diamond: true}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, ok := s.tableauPiles[1].CardBoundsAt(tt.cardIndex)
			if !ok {
				t.Fatalf("expected card %d in tableau pile 1", tt.cardIndex)
			}
			s.mediator.Handle(Event{Type: EventPointerDown, X: b.Min.X + 1, Y: b.Min.Y + 1})
			defer s.mediator.Handle(Event{Type: EventPointerUp, X: -1, Y: -1})

			got, ok := s.mediator.ValidDropTargets()
			if !ok {
				t.Fatal("expected drop targets while dragging")
			}
			if got != tt.want {
				t.Errorf("ValidDropTargets() = %+v; want %+v", got, tt.want)
			}
		})
	}
}

func TestSolitaireScene_SyncComponents_HighlightsDropTargets(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{
		Waste: []card.Card{{Suit: card.Spade, Rank: card.King}},
	})
	// A black King can go onto any empty tableau pile; pile 0 is not empty.
	s.board.Tableau[0] = solitaire.Pile{Cards: []card.Card{{Suit: card.Heart, Rank: card.Two}}, FaceUp: 1}
	s.syncComponents()

	for i, p := range s.tableauPiles {
		if len(p.Highlights()) != 0 {
			t.Errorf("tableau pile %d highlighted before any drag", i)
		}
	}

	top, _ := s.wastePile.TopBounds()
	s.mediator.Handle(Event{Type: EventPointerDown, X: top.Min.X, Y: top.Min.Y})
	s.syncComponents()

	if got := s.tableauPiles[0].Highlights(); len(got) != 0 {
		t.Errorf("tableau pile 0 highlights = %v; want none", got)
	}
	for i := 1; i < solitaire.TableauPileCount; i++ {
		x := tableauOriginX + i*tableauGapX
		want := image.Rect(x, tableauOriginY, x+cardWidth, tableauOriginY+cardHeight)
		if got := s.tableauPiles[i].Highlights(); len(got) != 1 || got[0] != want {
			t.Errorf("tableau pile %d highlights = %v; want the empty slot %v", i, got, want)
		}
	}
	if got := s.foundationPile.Highlights(); len(got) != 0 {
		t.Errorf("foundation highlights = %v; want none", got)
	}

	// Drawing with highlights set must not panic.
	s.Draw(ebiten.NewImage(1280, 720))

	s.mediator.Handle(Event{Type: EventPointerUp, X: -1, Y: -1})
	s.syncComponents()
	for i, p := range s.tableauPiles {
		if len(p.Highlights()) != 0 {
			t.Errorf("tableau pile %d still highlighted after the drag ended", i)
		}
	}
}

func TestSolitaireScene_SyncComponents_HighlightsFoundationSlotAndTopCard(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{
		Waste: []card.Card{{Suit: card.Club, Rank: card.Ace}},
	})
	s.board.Tableau[0] = solitaire.Pile{Cards: []card.Card{{Suit: card.Heart, Rank: card.Two}}, FaceUp: 1}
	s.syncComponents()

	top, _ := s.wastePile.TopBounds()
	s.mediator.Handle(Event{Type: EventPointerDown, X: top.Min.X, Y: top.Min.Y})
	s.syncComponents()

	// The black Ace fits on the red Two in pile 0 and on the Club
	// foundation; the outline should surround the Two itself and the
	// Club slot specifically.
	wantTop, _ := s.tableauPiles[0].TopBounds()
	if got := s.tableauPiles[0].Highlights(); len(got) != 1 || got[0] != wantTop {
		t.Errorf("tableau pile 0 highlights = %v; want its top card %v", got, wantTop)
	}
	x := foundationOriginX + int(card.Club)*foundationGapX
	wantSlot := image.Rect(x, foundationOriginY, x+cardWidth, foundationOriginY+cardHeight)
	if got := s.foundationPile.Highlights(); len(got) != 1 || got[0] != wantSlot {
		t.Errorf("foundation highlights = %v; want the Club slot %v", got, wantSlot)
	}
}
