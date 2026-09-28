package solitaire

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/haruki7049/trump-center/internal/card"
	"github.com/haruki7049/trump-center/internal/solitaire"
)

// newTestScene builds a SolitaireScene with its images and component tree
// (and Mediator) preloaded, so tests can freely replace its board with a
// known layout.
func newTestScene(t *testing.T) *SolitaireScene {
	t.Helper()

	s, err := NewSolitaireScene()
	if err != nil {
		t.Fatalf("failed to create SolitaireScene: %v", err)
	}
	return s
}

func TestMediator_StartDrag(t *testing.T) {
	s := newTestScene(t)

	c := card.Card{Suit: card.Heart, Rank: card.Queen}
	s.mediator.startDrag(dragSourceWaste, 3, c, 15, 25, 5, 10)

	got, ok := s.mediator.Dragging()
	if !ok {
		t.Fatal("expected a drag to be active")
	}
	want := DragState{
		Source:       dragSourceWaste,
		TableauIndex: 3,
		Card:         c,
		OffsetX:      10,
		OffsetY:      15,
	}
	if got != want {
		t.Errorf("Dragging() = %+v; want %+v", got, want)
	}
}

func TestMediator_HandlePointerDown(t *testing.T) {
	s := newTestScene(t)

	s.board = &solitaire.Board{
		Stock: []card.Card{{Suit: card.Spade, Rank: card.Ace}},
		Waste: []card.Card{{Suit: card.Heart, Rank: card.King}},
	}
	s.mediator.board = s.board
	s.board.Tableau[0] = solitaire.Pile{
		Cards:  []card.Card{{Suit: card.Club, Rank: card.Ten}},
		FaceUp: 1,
	}
	s.syncComponents()

	// Clicking the stock draws a card.
	stockPt := s.stockPile.Bounds().Min
	s.mediator.handlePointerDown(stockPt.X, stockPt.Y)
	if len(s.board.Stock) != 0 || len(s.board.Waste) != 2 {
		t.Fatalf("expected clicking the stock to draw a card, got stock=%d waste=%d", len(s.board.Stock), len(s.board.Waste))
	}

	// Clicking the top of the waste picks it up.
	s.syncComponents()
	wasteBounds, ok := s.wastePile.TopBounds()
	if !ok {
		t.Fatal("expected the waste to have a top card after drawing")
	}
	s.mediator.handlePointerDown(wasteBounds.Min.X, wasteBounds.Min.Y)
	if ds, ok := s.mediator.Dragging(); !ok || ds.Source != dragSourceWaste {
		t.Errorf("expected clicking the waste top to start a waste drag, got %+v (ok=%v)", ds, ok)
	}
	s.mediator.state = MediatorIdle
	s.mediator.drag = dragInfo{}

	// Clicking the top of a tableau pile picks it up.
	s.syncComponents()
	tabBounds, ok := s.tableauPiles[0].TopBounds()
	if !ok {
		t.Fatal("expected tableau pile 0 to have a top card")
	}
	s.mediator.handlePointerDown(tabBounds.Min.X, tabBounds.Min.Y)
	if ds, ok := s.mediator.Dragging(); !ok || ds.Source != dragSourceTableau || ds.TableauIndex != 0 {
		t.Errorf("expected clicking the tableau top to start a tableau drag, got %+v (ok=%v)", ds, ok)
	}
	s.mediator.state = MediatorIdle
	s.mediator.drag = dragInfo{}

	// A face-down tableau top card cannot be picked up.
	s.board.Tableau[1] = solitaire.Pile{
		Cards:  []card.Card{{Suit: card.Diamond, Rank: card.Five}},
		FaceUp: 0,
	}
	s.mediator.handlePointerDown(tableauOriginX+1*tableauGapX, tableauOriginY)
	if _, ok := s.mediator.Dragging(); ok {
		t.Error("expected clicking a face-down tableau card to not start a drag")
	}

	// Clicking empty space does nothing.
	s.mediator.handlePointerDown(-100, -100)
	if _, ok := s.mediator.Dragging(); ok {
		t.Error("expected clicking empty space to not start a drag")
	}
}

func TestMediator_HandlePointerUp(t *testing.T) {
	t.Run("waste to foundation", func(t *testing.T) {
		s := newTestScene(t)
		s.board = &solitaire.Board{Waste: []card.Card{{Suit: card.Spade, Rank: card.Ace}}}
		s.mediator.board = s.board
		s.mediator.startDrag(dragSourceWaste, -1, card.Card{Suit: card.Spade, Rank: card.Ace}, 0, 0, 0, 0)
		s.syncComponents()

		pt := s.foundationPile.Bounds().Min
		s.mediator.handlePointerUp(pt.X, pt.Y)

		if len(s.board.Waste) != 0 || len(s.board.Foundation[card.Spade]) != 1 {
			t.Errorf("expected the Ace to move to the foundation, got waste=%d foundation=%d", len(s.board.Waste), len(s.board.Foundation[card.Spade]))
		}
		if _, ok := s.mediator.Dragging(); ok {
			t.Error("expected the drag to end after dropping")
		}
	})

	t.Run("waste to tableau", func(t *testing.T) {
		s := newTestScene(t)
		s.board = &solitaire.Board{Waste: []card.Card{{Suit: card.Heart, Rank: card.Queen}}}
		s.mediator.board = s.board
		s.board.Tableau[2] = solitaire.Pile{Cards: []card.Card{{Suit: card.Spade, Rank: card.King}}, FaceUp: 1}
		s.mediator.startDrag(dragSourceWaste, -1, card.Card{Suit: card.Heart, Rank: card.Queen}, 0, 0, 0, 0)
		s.syncComponents()

		pt, ok := s.tableauPiles[2].TopBounds()
		if !ok {
			t.Fatal("expected tableau pile 2 to have a top card")
		}
		s.mediator.handlePointerUp(pt.Min.X, pt.Min.Y)

		if len(s.board.Waste) != 0 || len(s.board.Tableau[2].Cards) != 2 {
			t.Errorf("expected the Queen to move onto the King, got waste=%d tableau=%d", len(s.board.Waste), len(s.board.Tableau[2].Cards))
		}
	})

	t.Run("tableau to foundation", func(t *testing.T) {
		s := newTestScene(t)
		s.board = &solitaire.Board{}
		s.mediator.board = s.board
		s.board.Tableau[0] = solitaire.Pile{Cards: []card.Card{{Suit: card.Club, Rank: card.Ace}}, FaceUp: 1}
		s.mediator.startDrag(dragSourceTableau, 0, card.Card{Suit: card.Club, Rank: card.Ace}, 0, 0, 0, 0)
		s.syncComponents()

		pt := s.foundationPile.Bounds().Min
		s.mediator.handlePointerUp(pt.X, pt.Y)

		if len(s.board.Tableau[0].Cards) != 0 || len(s.board.Foundation[card.Club]) != 1 {
			t.Errorf("expected the Ace to move to the foundation, got tableau=%d foundation=%d", len(s.board.Tableau[0].Cards), len(s.board.Foundation[card.Club]))
		}
	})

	t.Run("tableau to tableau", func(t *testing.T) {
		s := newTestScene(t)
		s.board = &solitaire.Board{}
		s.mediator.board = s.board
		s.board.Tableau[0] = solitaire.Pile{Cards: []card.Card{{Suit: card.Heart, Rank: card.Ten}}, FaceUp: 1}
		s.board.Tableau[1] = solitaire.Pile{Cards: []card.Card{{Suit: card.Spade, Rank: card.Jack}}, FaceUp: 1}
		s.mediator.startDrag(dragSourceTableau, 0, card.Card{Suit: card.Heart, Rank: card.Ten}, 0, 0, 0, 0)
		s.syncComponents()

		pt, ok := s.tableauPiles[1].TopBounds()
		if !ok {
			t.Fatal("expected tableau pile 1 to have a top card")
		}
		s.mediator.handlePointerUp(pt.Min.X, pt.Min.Y)

		if len(s.board.Tableau[0].Cards) != 0 || len(s.board.Tableau[1].Cards) != 2 {
			t.Errorf("expected the Ten to move onto the Jack, got source=%d dest=%d", len(s.board.Tableau[0].Cards), len(s.board.Tableau[1].Cards))
		}
	})

	t.Run("dropping onto the same tableau pile does nothing", func(t *testing.T) {
		s := newTestScene(t)
		s.board = &solitaire.Board{}
		s.mediator.board = s.board
		s.board.Tableau[0] = solitaire.Pile{Cards: []card.Card{{Suit: card.Heart, Rank: card.Ten}}, FaceUp: 1}
		s.mediator.startDrag(dragSourceTableau, 0, card.Card{Suit: card.Heart, Rank: card.Ten}, 0, 0, 0, 0)
		s.syncComponents()

		pt := s.tableauPiles[0].Bounds().Min
		s.mediator.handlePointerUp(pt.X, pt.Y)

		if len(s.board.Tableau[0].Cards) != 1 {
			t.Errorf("expected dropping a pile onto itself to be a no-op, got %d cards", len(s.board.Tableau[0].Cards))
		}
	})

	t.Run("drop on empty space does nothing but still ends the drag", func(t *testing.T) {
		s := newTestScene(t)
		s.board = &solitaire.Board{Waste: []card.Card{{Suit: card.Spade, Rank: card.Ace}}}
		s.mediator.board = s.board
		s.mediator.startDrag(dragSourceWaste, -1, card.Card{Suit: card.Spade, Rank: card.Ace}, 0, 0, 0, 0)
		s.syncComponents()

		s.mediator.handlePointerUp(-100, -100)

		if len(s.board.Waste) != 1 {
			t.Errorf("expected the drop to be ignored, got waste=%d", len(s.board.Waste))
		}
		if _, ok := s.mediator.Dragging(); ok {
			t.Error("expected the drag to end even when the drop location is empty")
		}
	})

	t.Run("does nothing when no drag is in progress", func(t *testing.T) {
		s := newTestScene(t)
		s.board = &solitaire.Board{Waste: []card.Card{{Suit: card.Spade, Rank: card.Ace}}}
		s.mediator.board = s.board
		s.syncComponents()

		pt := s.foundationPile.Bounds().Min
		s.mediator.handlePointerUp(pt.X, pt.Y)

		if len(s.board.Waste) != 1 {
			t.Errorf("expected no move without an active drag, got waste=%d", len(s.board.Waste))
		}
	})
}

// TestMediator_WiredAsRootEventHandler verifies the full chain end to
// end: NewSolitaireScene installs the Mediator as the RootComponent's
// event handler, so dispatching a real Event at a hit-tested leaf (not
// calling any Mediator method directly) reaches it and acts on the
// board, exactly as SolitaireScene.Update does.
func TestMediator_WiredAsRootEventHandler(t *testing.T) {
	s := newTestScene(t)
	s.board = &solitaire.Board{
		Stock: []card.Card{{Suit: card.Spade, Rank: card.Ace}},
	}
	s.mediator.board = s.board
	s.syncComponents()

	pt := s.stockPile.Bounds().Min
	hit := s.root.HitTest(pt.X, pt.Y)
	if hit == nil {
		t.Fatal("expected the stock pile to be hit-tested at its own bounds")
	}

	if handled := Dispatch(hit, Event{Type: EventPointerDown, X: pt.X, Y: pt.Y}); !handled {
		t.Fatal("expected the event to be handled once it reaches the Mediator at Root")
	}

	if len(s.board.Stock) != 0 || len(s.board.Waste) != 1 {
		t.Errorf("expected the dispatched event to draw a card via the Mediator, got stock=%d waste=%d", len(s.board.Stock), len(s.board.Waste))
	}
}

func TestSolitaireScene_Draw_WithActiveDrag(t *testing.T) {
	s := newTestScene(t)
	s.board = &solitaire.Board{}
	s.mediator.board = s.board
	s.mediator.startDrag(dragSourceNone, -1, card.Card{Suit: card.Spade, Rank: card.Ace}, 0, 0, 0, 0)

	screen := ebiten.NewImage(1280, 720)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Draw panicked: %v", r)
		}
	}()

	s.Draw(screen)
}
