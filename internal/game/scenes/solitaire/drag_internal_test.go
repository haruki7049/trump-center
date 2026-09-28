package solitaire

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/haruki7049/trump-center/internal/card"
	"github.com/haruki7049/trump-center/internal/solitaire"
)

// newTestScene builds a SolitaireScene with its images and component tree
// preloaded, so tests can freely replace its board with a known layout.
func newTestScene(t *testing.T) *SolitaireScene {
	t.Helper()

	s, err := NewSolitaireScene()
	if err != nil {
		t.Fatalf("failed to create SolitaireScene: %v", err)
	}
	return s
}

func TestSolitaireScene_StartDrag(t *testing.T) {
	s := newTestScene(t)

	c := card.Card{Suit: card.Heart, Rank: card.Queen}
	s.startDrag(dragSourceWaste, 3, c, 15, 25, 5, 10)

	want := drag{
		active:       true,
		source:       dragSourceWaste,
		tableauIndex: 3,
		card:         c,
		offsetX:      10,
		offsetY:      15,
		cursorX:      15,
		cursorY:      25,
	}
	if s.drag != want {
		t.Errorf("drag = %+v; want %+v", s.drag, want)
	}
}

func TestSolitaireScene_TryStartDrag(t *testing.T) {
	s := newTestScene(t)

	s.board = &solitaire.Board{
		Stock: []card.Card{{Suit: card.Spade, Rank: card.Ace}},
		Waste: []card.Card{{Suit: card.Heart, Rank: card.King}},
	}
	s.board.Tableau[0] = solitaire.Pile{
		Cards:  []card.Card{{Suit: card.Club, Rank: card.Ten}},
		FaceUp: 1,
	}
	s.syncComponents()

	// Clicking the stock draws a card.
	stockPt := s.stockPile.Bounds().Min
	s.tryStartDrag(stockPt.X, stockPt.Y)
	if len(s.board.Stock) != 0 || len(s.board.Waste) != 2 {
		t.Fatalf("expected clicking the stock to draw a card, got stock=%d waste=%d", len(s.board.Stock), len(s.board.Waste))
	}

	// Clicking the top of the waste picks it up.
	s.syncComponents()
	wasteBounds, ok := s.wastePile.TopBounds()
	if !ok {
		t.Fatal("expected the waste to have a top card after drawing")
	}
	s.tryStartDrag(wasteBounds.Min.X, wasteBounds.Min.Y)
	if !s.drag.active || s.drag.source != dragSourceWaste {
		t.Errorf("expected clicking the waste top to start a waste drag, got %+v", s.drag)
	}
	s.drag = drag{}

	// Clicking the top of a tableau pile picks it up.
	s.syncComponents()
	tabBounds, ok := s.tableauPiles[0].TopBounds()
	if !ok {
		t.Fatal("expected tableau pile 0 to have a top card")
	}
	s.tryStartDrag(tabBounds.Min.X, tabBounds.Min.Y)
	if !s.drag.active || s.drag.source != dragSourceTableau || s.drag.tableauIndex != 0 {
		t.Errorf("expected clicking the tableau top to start a tableau drag, got %+v", s.drag)
	}
	s.drag = drag{}

	// A face-down tableau top card cannot be picked up.
	s.board.Tableau[1] = solitaire.Pile{
		Cards:  []card.Card{{Suit: card.Diamond, Rank: card.Five}},
		FaceUp: 0,
	}
	s.tryStartDrag(tableauOriginX+1*tableauGapX, tableauOriginY)
	if s.drag.active {
		t.Errorf("expected clicking a face-down tableau card to not start a drag, got %+v", s.drag)
	}

	// Clicking empty space does nothing.
	s.tryStartDrag(-100, -100)
	if s.drag.active {
		t.Errorf("expected clicking empty space to not start a drag, got %+v", s.drag)
	}
}

func TestSolitaireScene_DropDrag(t *testing.T) {
	t.Run("waste to foundation", func(t *testing.T) {
		s := newTestScene(t)
		s.board = &solitaire.Board{Waste: []card.Card{{Suit: card.Spade, Rank: card.Ace}}}
		s.drag = drag{active: true, source: dragSourceWaste}
		s.syncComponents()

		pt := s.foundationPile.Bounds().Min
		s.dropDrag(pt.X, pt.Y)

		if len(s.board.Waste) != 0 || len(s.board.Foundation[card.Spade]) != 1 {
			t.Errorf("expected the Ace to move to the foundation, got waste=%d foundation=%d", len(s.board.Waste), len(s.board.Foundation[card.Spade]))
		}
	})

	t.Run("waste to tableau", func(t *testing.T) {
		s := newTestScene(t)
		s.board = &solitaire.Board{Waste: []card.Card{{Suit: card.Heart, Rank: card.Queen}}}
		s.board.Tableau[2] = solitaire.Pile{Cards: []card.Card{{Suit: card.Spade, Rank: card.King}}, FaceUp: 1}
		s.drag = drag{active: true, source: dragSourceWaste}
		s.syncComponents()

		pt, ok := s.tableauPiles[2].TopBounds()
		if !ok {
			t.Fatal("expected tableau pile 2 to have a top card")
		}
		s.dropDrag(pt.Min.X, pt.Min.Y)

		if len(s.board.Waste) != 0 || len(s.board.Tableau[2].Cards) != 2 {
			t.Errorf("expected the Queen to move onto the King, got waste=%d tableau=%d", len(s.board.Waste), len(s.board.Tableau[2].Cards))
		}
	})

	t.Run("tableau to foundation", func(t *testing.T) {
		s := newTestScene(t)
		s.board = &solitaire.Board{}
		s.board.Tableau[0] = solitaire.Pile{Cards: []card.Card{{Suit: card.Club, Rank: card.Ace}}, FaceUp: 1}
		s.drag = drag{active: true, source: dragSourceTableau, tableauIndex: 0}
		s.syncComponents()

		pt := s.foundationPile.Bounds().Min
		s.dropDrag(pt.X, pt.Y)

		if len(s.board.Tableau[0].Cards) != 0 || len(s.board.Foundation[card.Club]) != 1 {
			t.Errorf("expected the Ace to move to the foundation, got tableau=%d foundation=%d", len(s.board.Tableau[0].Cards), len(s.board.Foundation[card.Club]))
		}
	})

	t.Run("tableau to tableau", func(t *testing.T) {
		s := newTestScene(t)
		s.board = &solitaire.Board{}
		s.board.Tableau[0] = solitaire.Pile{Cards: []card.Card{{Suit: card.Heart, Rank: card.Ten}}, FaceUp: 1}
		s.board.Tableau[1] = solitaire.Pile{Cards: []card.Card{{Suit: card.Spade, Rank: card.Jack}}, FaceUp: 1}
		s.drag = drag{active: true, source: dragSourceTableau, tableauIndex: 0}
		s.syncComponents()

		pt, ok := s.tableauPiles[1].TopBounds()
		if !ok {
			t.Fatal("expected tableau pile 1 to have a top card")
		}
		s.dropDrag(pt.Min.X, pt.Min.Y)

		if len(s.board.Tableau[0].Cards) != 0 || len(s.board.Tableau[1].Cards) != 2 {
			t.Errorf("expected the Ten to move onto the Jack, got source=%d dest=%d", len(s.board.Tableau[0].Cards), len(s.board.Tableau[1].Cards))
		}
	})

	t.Run("dropping onto the same tableau pile does nothing", func(t *testing.T) {
		s := newTestScene(t)
		s.board = &solitaire.Board{}
		s.board.Tableau[0] = solitaire.Pile{Cards: []card.Card{{Suit: card.Heart, Rank: card.Ten}}, FaceUp: 1}
		s.drag = drag{active: true, source: dragSourceTableau, tableauIndex: 0}
		s.syncComponents()

		pt := s.tableauPiles[0].Bounds().Min
		s.dropDrag(pt.X, pt.Y)

		if len(s.board.Tableau[0].Cards) != 1 {
			t.Errorf("expected dropping a pile onto itself to be a no-op, got %d cards", len(s.board.Tableau[0].Cards))
		}
	})

	t.Run("drop on empty space does nothing", func(t *testing.T) {
		s := newTestScene(t)
		s.board = &solitaire.Board{Waste: []card.Card{{Suit: card.Spade, Rank: card.Ace}}}
		s.drag = drag{active: true, source: dragSourceWaste}
		s.syncComponents()

		s.dropDrag(-100, -100)

		if len(s.board.Waste) != 1 {
			t.Errorf("expected the drop to be ignored, got waste=%d", len(s.board.Waste))
		}
	})
}

func TestSolitaireScene_Draw_WithActiveDrag(t *testing.T) {
	s := newTestScene(t)
	s.board = &solitaire.Board{}
	s.drag = drag{
		active:  true,
		card:    card.Card{Suit: card.Spade, Rank: card.Ace},
		cursorX: 100,
		cursorY: 100,
	}

	screen := ebiten.NewImage(1280, 720)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Draw panicked: %v", r)
		}
	}()

	s.Draw(screen)
}
