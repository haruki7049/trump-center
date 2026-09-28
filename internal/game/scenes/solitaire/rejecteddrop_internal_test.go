package solitaire

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/haruki7049/trump-center/internal/card"
	"github.com/haruki7049/trump-center/internal/solitaire"
)

// newRejectionTestScene builds a scene where the waste top (the Seven of
// Hearts) fits nowhere: tableau pile 0's top is a same-color Eight, and
// its foundation is empty. Tableau pile 1 holds a black Eight it does fit
// on, for the successful-drop case.
func newRejectionTestScene(t *testing.T) *SolitaireScene {
	t.Helper()

	s := newTestScene(t)
	s.setBoard(&solitaire.Board{
		Waste: []card.Card{{Suit: card.Heart, Rank: card.Seven}},
	})
	s.board.Tableau[0] = solitaire.Pile{Cards: []card.Card{{Suit: card.Diamond, Rank: card.Eight}}, FaceUp: 1}
	s.board.Tableau[1] = solitaire.Pile{Cards: []card.Card{{Suit: card.Club, Rank: card.Eight}}, FaceUp: 1}
	s.board.Tableau[2] = solitaire.Pile{Cards: []card.Card{
		{Suit: card.Spade, Rank: card.Two},
		{Suit: card.Heart, Rank: card.Ace},
	}, FaceUp: 2}
	s.syncComponents()
	return s
}

func (s *SolitaireScene) dragWasteTopTo(x, y int) {
	top, _ := s.wastePile.TopBounds()
	s.mediator.Handle(Event{Type: EventPointerDown, X: top.Min.X, Y: top.Min.Y})
	s.mediator.Handle(Event{Type: EventPointerUp, X: x, Y: y})
}

func TestMediator_RejectedDrop(t *testing.T) {
	foundationPt := image.Pt(foundationOriginX+1, foundationOriginY+1)
	tableauPt := func(i int) image.Point {
		return image.Pt(tableauOriginX+i*tableauGapX+1, tableauOriginY+1)
	}

	tests := []struct {
		name   string
		drop   func(s *SolitaireScene)
		want   RejectedDrop
		wantOK bool
	}{
		{
			name:   "illegal tableau drop",
			drop:   func(s *SolitaireScene) { p := tableauPt(0); s.dragWasteTopTo(p.X, p.Y) },
			want:   RejectedDrop{Target: DropTargetTableau, Index: 0},
			wantOK: true,
		},
		{
			name:   "illegal foundation drop is reported on the card's own suit",
			drop:   func(s *SolitaireScene) { s.dragWasteTopTo(foundationPt.X, foundationPt.Y) },
			want:   RejectedDrop{Target: DropTargetFoundation, Index: int(card.Heart)},
			wantOK: true,
		},
		{
			name: "multi-card run onto the foundation",
			drop: func(s *SolitaireScene) {
				// The Two-Ace run: the Ace alone would be legal, but a run
				// never goes to the foundation.
				b, _ := s.tableauPiles[2].CardBoundsAt(0)
				s.mediator.Handle(Event{Type: EventPointerDown, X: b.Min.X + 1, Y: b.Min.Y + 1})
				s.mediator.Handle(Event{Type: EventPointerUp, X: foundationPt.X, Y: foundationPt.Y})
			},
			want:   RejectedDrop{Target: DropTargetFoundation, Index: int(card.Spade)},
			wantOK: true,
		},
		{
			name:   "legal drop",
			drop:   func(s *SolitaireScene) { p := tableauPt(1); s.dragWasteTopTo(p.X, p.Y) },
			wantOK: false,
		},
		{
			name:   "drop on empty space just cancels",
			drop:   func(s *SolitaireScene) { s.dragWasteTopTo(-1, -1) },
			wantOK: false,
		},
		{
			name: "drop back on the source pile just cancels",
			drop: func(s *SolitaireScene) {
				top, _ := s.tableauPiles[0].TopBounds()
				s.mediator.Handle(Event{Type: EventPointerDown, X: top.Min.X + 1, Y: top.Min.Y + 1})
				s.mediator.Handle(Event{Type: EventPointerUp, X: top.Min.X + 1, Y: top.Min.Y + 1})
			},
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newRejectionTestScene(t)
			tt.drop(s)

			got, ok := s.mediator.RejectedDrop()
			if ok != tt.wantOK || got != tt.want {
				t.Errorf("RejectedDrop() = %+v, %v; want %+v, %v", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestMediator_RejectedDrop_ExpiresAfterTicks(t *testing.T) {
	s := newRejectionTestScene(t)
	p := image.Pt(tableauOriginX+1, tableauOriginY+1)
	s.dragWasteTopTo(p.X, p.Y)

	for range rejectedDropTicks - 1 {
		s.mediator.Tick()
	}
	if _, ok := s.mediator.RejectedDrop(); !ok {
		t.Fatal("expected the rejection to still be reported one tick before it expires")
	}
	s.mediator.Tick()
	if _, ok := s.mediator.RejectedDrop(); ok {
		t.Error("expected the rejection to expire after rejectedDropTicks ticks")
	}
}

func TestMediator_RejectedDrop_ClearedByNextClick(t *testing.T) {
	s := newRejectionTestScene(t)
	p := image.Pt(tableauOriginX+1, tableauOriginY+1)
	s.dragWasteTopTo(p.X, p.Y)

	s.mediator.Handle(Event{Type: EventPointerDown, X: -1, Y: -1})
	if _, ok := s.mediator.RejectedDrop(); ok {
		t.Error("expected a new click to clear the rejection")
	}
}

func TestSolitaireScene_SyncComponents_HighlightsRejectedDrop(t *testing.T) {
	s := newRejectionTestScene(t)
	p := image.Pt(tableauOriginX+1, tableauOriginY+1)
	s.dragWasteTopTo(p.X, p.Y)
	s.syncComponents()

	wantRect, _ := s.tableauPiles[0].TopBounds()
	want := Highlight{Rect: wantRect, Color: rejectedDropColor}
	if got := s.tableauPiles[0].Highlights(); len(got) != 1 || got[0] != want {
		t.Errorf("tableau pile 0 highlights = %v; want %v", got, want)
	}
	for i := 1; i < solitaire.TableauPileCount; i++ {
		if got := s.tableauPiles[i].Highlights(); len(got) != 0 {
			t.Errorf("tableau pile %d highlights = %v; want none", i, got)
		}
	}

	// Drawing with a rejection outline set must not panic.
	s.Draw(ebiten.NewImage(1280, 720))

	for range rejectedDropTicks {
		if _, err := s.Update(); err != nil {
			t.Fatalf("unexpected error from Update: %v", err)
		}
	}
	s.syncComponents()
	if got := s.tableauPiles[0].Highlights(); len(got) != 0 {
		t.Errorf("tableau pile 0 highlights = %v after the rejection expired; want none", got)
	}
}

func TestSolitaireScene_SyncComponents_HighlightsRejectedFoundationSlot(t *testing.T) {
	s := newRejectionTestScene(t)
	s.dragWasteTopTo(foundationOriginX+1, foundationOriginY+1)
	s.syncComponents()

	x := foundationOriginX + int(card.Heart)*foundationGapX
	want := Highlight{
		Rect:  image.Rect(x, foundationOriginY, x+cardWidth, foundationOriginY+cardHeight),
		Color: rejectedDropColor,
	}
	if got := s.foundationPile.Highlights(); len(got) != 1 || got[0] != want {
		t.Errorf("foundation highlights = %v; want the Heart slot %v", got, want)
	}
}
