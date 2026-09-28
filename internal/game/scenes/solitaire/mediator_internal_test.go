package solitaire

import (
	"slices"
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

// setBoard replaces s's board with board and points the Mediator at it,
// keeping both in sync the way NewSolitaireScene does at construction
// (including restarting the background Stuck check).
func (s *SolitaireScene) setBoard(board *solitaire.Board) {
	s.board = board
	s.mediator.board = board
	s.mediator.refreshStuck()
}

// waitStuck blocks until the Mediator's background Stuck check (if any)
// finishes, and records its result the way Tick would.
func (m *Mediator) waitStuck() {
	if m.stuckResult != nil {
		m.stuck = <-m.stuckResult
		m.stuckResult = nil
	}
}

func TestMediator_Handle_PointerDown_DrawsFromStock(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{Stock: []card.Card{{Suit: card.Spade, Rank: card.Ace}}})
	s.syncComponents()

	pt := s.stockPile.Bounds().Min
	s.mediator.Handle(Event{Type: EventPointerDown, X: pt.X, Y: pt.Y})

	if len(s.board.Stock) != 0 || len(s.board.Waste) != 1 {
		t.Errorf("expected clicking the stock to draw a card, got stock=%d waste=%d", len(s.board.Stock), len(s.board.Waste))
	}
	if _, ok := s.mediator.Dragging(); ok {
		t.Error("expected drawing from the stock to not start a drag")
	}
}

func TestMediator_Handle_PointerDown_PicksUpWasteTop(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{Waste: []card.Card{{Suit: card.Heart, Rank: card.King}}})
	s.syncComponents()

	pt, ok := s.wastePile.TopBounds()
	if !ok {
		t.Fatal("expected the waste to have a top card")
	}
	s.mediator.Handle(Event{Type: EventPointerDown, X: pt.Min.X, Y: pt.Min.Y})

	got, ok := s.mediator.Dragging()
	if !ok {
		t.Fatal("expected clicking the waste top to start a drag")
	}
	want := []card.Card{{Suit: card.Heart, Rank: card.King}}
	if got.Source != dragSourceWaste || !slices.Equal(got.Cards, want) {
		t.Errorf("Dragging() = %+v; want the King of Hearts from the waste", got)
	}
}

func TestMediator_Handle_PointerDown_PicksUpTableauTop(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{})
	s.board.Tableau[0] = solitaire.Pile{
		Cards:  []card.Card{{Suit: card.Club, Rank: card.Ten}},
		FaceUp: 1,
	}
	s.syncComponents()

	pt, ok := s.tableauPiles[0].TopBounds()
	if !ok {
		t.Fatal("expected tableau pile 0 to have a top card")
	}
	s.mediator.Handle(Event{Type: EventPointerDown, X: pt.Min.X, Y: pt.Min.Y})

	got, ok := s.mediator.Dragging()
	if !ok {
		t.Fatal("expected clicking the tableau top to start a drag")
	}
	if got.Source != dragSourceTableau || got.TableauIndex != 0 {
		t.Errorf("Dragging() = %+v; want source=tableau, tableauIndex=0", got)
	}
}

// TestMediator_Handle_PointerDown_PicksUpMiddleOfRun verifies that
// clicking a face-up card that isn't the pile's top picks up that card
// and every card above it as a run, not just the single top card.
func TestMediator_Handle_PointerDown_PicksUpMiddleOfRun(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{})
	run := []card.Card{
		{Suit: card.Club, Rank: card.Queen},
		{Suit: card.Heart, Rank: card.Jack},
		{Suit: card.Club, Rank: card.Ten},
	}
	s.board.Tableau[0] = solitaire.Pile{
		Cards:  append([]card.Card{{Suit: card.Diamond, Rank: card.King}}, run...),
		FaceUp: 3,
	}
	s.syncComponents()

	// Click the Jack (index 2 in the pile, the middle of the run), not
	// the top (the Ten, index 3).
	b, ok := s.tableauPiles[0].CardBoundsAt(2)
	if !ok {
		t.Fatal("expected the Jack to have bounds")
	}
	s.mediator.Handle(Event{Type: EventPointerDown, X: b.Min.X, Y: b.Min.Y})

	got, ok := s.mediator.Dragging()
	if !ok {
		t.Fatal("expected clicking the Jack to start a drag")
	}
	want := []card.Card{
		{Suit: card.Heart, Rank: card.Jack},
		{Suit: card.Club, Rank: card.Ten},
	}
	if got.Source != dragSourceTableau || got.TableauIndex != 0 || got.CardIndex != 2 || !slices.Equal(got.Cards, want) {
		t.Errorf("Dragging() = %+v; want source=tableau, tableauIndex=0, cardIndex=2, cards=%+v", got, want)
	}
}

func TestMediator_Handle_PointerDown_FaceDownTableauTop_NoOp(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{})
	s.board.Tableau[0] = solitaire.Pile{
		Cards:  []card.Card{{Suit: card.Diamond, Rank: card.Five}},
		FaceUp: 0,
	}
	s.syncComponents()

	s.mediator.Handle(Event{Type: EventPointerDown, X: tableauOriginX, Y: tableauOriginY})

	if _, ok := s.mediator.Dragging(); ok {
		t.Error("expected clicking a face-down tableau card to not start a drag")
	}
}

func TestMediator_Handle_PointerDown_EmptySpace_NoOp(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{})
	s.syncComponents()

	s.mediator.Handle(Event{Type: EventPointerDown, X: -100, Y: -100})

	if _, ok := s.mediator.Dragging(); ok {
		t.Error("expected clicking empty space to not start a drag")
	}
}

func TestMediator_Handle_PointerDown_WhileAlreadyDragging_NoOp(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{
		Waste: []card.Card{{Suit: card.Heart, Rank: card.King}},
	})
	s.board.Tableau[0] = solitaire.Pile{
		Cards:  []card.Card{{Suit: card.Club, Rank: card.Ten}},
		FaceUp: 1,
	}
	s.syncComponents()

	wasteTop, ok := s.wastePile.TopBounds()
	if !ok {
		t.Fatal("expected the waste to have a top card")
	}
	s.mediator.Handle(Event{Type: EventPointerDown, X: wasteTop.Min.X, Y: wasteTop.Min.Y})

	before, ok := s.mediator.Dragging()
	if !ok {
		t.Fatal("expected the first pointer down to start a drag")
	}

	// A second pointer down, over a different pile, must not replace the
	// card already being dragged.
	s.syncComponents()
	tabTop, ok := s.tableauPiles[0].TopBounds()
	if !ok {
		t.Fatal("expected tableau pile 0 to have a top card")
	}
	s.mediator.Handle(Event{Type: EventPointerDown, X: tabTop.Min.X, Y: tabTop.Min.Y})

	after, ok := s.mediator.Dragging()
	if !ok {
		t.Fatal("expected the drag to still be active")
	}
	if after.Source != before.Source || after.TableauIndex != before.TableauIndex ||
		after.CardIndex != before.CardIndex || after.OffsetX != before.OffsetX ||
		after.OffsetY != before.OffsetY || !slices.Equal(after.Cards, before.Cards) {
		t.Errorf("expected the in-progress drag to be unaffected, got %+v; want %+v", after, before)
	}
}

func TestMediator_Handle_PointerUp_WithoutDrag_NoOp(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{Waste: []card.Card{{Suit: card.Spade, Rank: card.Ace}}})
	s.syncComponents()

	pt := s.foundationPile.Bounds().Min
	s.mediator.Handle(Event{Type: EventPointerUp, X: pt.X, Y: pt.Y})

	if len(s.board.Waste) != 1 {
		t.Errorf("expected no move without an active drag, got waste=%d", len(s.board.Waste))
	}
}

// TestMediator_Handle_DragAndDrop drives whole drag-and-drop interactions
// purely through synthetic Events dispatched at Mediator.Handle, the same
// entry point Dispatch uses once an event reaches Root: a pointer-down
// event at the source card's position, followed by a pointer-up event at
// the drop target's position.
func TestMediator_Handle_DragAndDrop(t *testing.T) {
	tests := []struct {
		name        string
		board       *solitaire.Board
		downAt      func(s *SolitaireScene) (x, y int)
		upAt        func(s *SolitaireScene) (x, y int)
		wantBoard   func(t *testing.T, b *solitaire.Board)
		wantDragged bool
	}{
		{
			name:  "waste to foundation",
			board: &solitaire.Board{Waste: []card.Card{{Suit: card.Spade, Rank: card.Ace}}},
			downAt: func(s *SolitaireScene) (int, int) {
				b, _ := s.wastePile.TopBounds()
				return b.Min.X, b.Min.Y
			},
			upAt: func(s *SolitaireScene) (int, int) {
				b := s.foundationPile.Bounds()
				return b.Min.X, b.Min.Y
			},
			wantBoard: func(t *testing.T, b *solitaire.Board) {
				t.Helper()
				if len(b.Waste) != 0 || len(b.Foundation[card.Spade]) != 1 {
					t.Errorf("expected the Ace to move to the foundation, got waste=%d foundation=%d", len(b.Waste), len(b.Foundation[card.Spade]))
				}
			},
		},
		{
			name: "waste to tableau",
			board: &solitaire.Board{
				Waste: []card.Card{{Suit: card.Heart, Rank: card.Queen}},
				Tableau: [solitaire.TableauPileCount]solitaire.Pile{
					2: {Cards: []card.Card{{Suit: card.Spade, Rank: card.King}}, FaceUp: 1},
				},
			},
			downAt: func(s *SolitaireScene) (int, int) {
				b, _ := s.wastePile.TopBounds()
				return b.Min.X, b.Min.Y
			},
			upAt: func(s *SolitaireScene) (int, int) {
				b, _ := s.tableauPiles[2].TopBounds()
				return b.Min.X, b.Min.Y
			},
			wantBoard: func(t *testing.T, b *solitaire.Board) {
				t.Helper()
				if len(b.Waste) != 0 || len(b.Tableau[2].Cards) != 2 {
					t.Errorf("expected the Queen to move onto the King, got waste=%d tableau=%d", len(b.Waste), len(b.Tableau[2].Cards))
				}
			},
		},
		{
			name: "tableau to foundation",
			board: &solitaire.Board{
				Tableau: [solitaire.TableauPileCount]solitaire.Pile{
					0: {Cards: []card.Card{{Suit: card.Club, Rank: card.Ace}}, FaceUp: 1},
				},
			},
			downAt: func(s *SolitaireScene) (int, int) {
				b, _ := s.tableauPiles[0].TopBounds()
				return b.Min.X, b.Min.Y
			},
			upAt: func(s *SolitaireScene) (int, int) {
				b := s.foundationPile.Bounds()
				return b.Min.X, b.Min.Y
			},
			wantBoard: func(t *testing.T, b *solitaire.Board) {
				t.Helper()
				if len(b.Tableau[0].Cards) != 0 || len(b.Foundation[card.Club]) != 1 {
					t.Errorf("expected the Ace to move to the foundation, got tableau=%d foundation=%d", len(b.Tableau[0].Cards), len(b.Foundation[card.Club]))
				}
			},
		},
		{
			name: "tableau to tableau",
			board: &solitaire.Board{
				Tableau: [solitaire.TableauPileCount]solitaire.Pile{
					0: {Cards: []card.Card{{Suit: card.Heart, Rank: card.Ten}}, FaceUp: 1},
					1: {Cards: []card.Card{{Suit: card.Spade, Rank: card.Jack}}, FaceUp: 1},
				},
			},
			downAt: func(s *SolitaireScene) (int, int) {
				b, _ := s.tableauPiles[0].TopBounds()
				return b.Min.X, b.Min.Y
			},
			upAt: func(s *SolitaireScene) (int, int) {
				b, _ := s.tableauPiles[1].TopBounds()
				return b.Min.X, b.Min.Y
			},
			wantBoard: func(t *testing.T, b *solitaire.Board) {
				t.Helper()
				if len(b.Tableau[0].Cards) != 0 || len(b.Tableau[1].Cards) != 2 {
					t.Errorf("expected the Ten to move onto the Jack, got source=%d dest=%d", len(b.Tableau[0].Cards), len(b.Tableau[1].Cards))
				}
			},
		},
		{
			name: "tableau run to tableau",
			board: &solitaire.Board{
				Tableau: [solitaire.TableauPileCount]solitaire.Pile{
					0: {Cards: []card.Card{
						{Suit: card.Spade, Rank: card.Nine}, // face-down
						{Suit: card.Heart, Rank: card.Eight},
						{Suit: card.Club, Rank: card.Seven},
					}, FaceUp: 2},
					1: {Cards: []card.Card{{Suit: card.Club, Rank: card.Nine}}, FaceUp: 1},
				},
			},
			downAt: func(s *SolitaireScene) (int, int) {
				// Index 1 is the Eight of Hearts, the bottom of the
				// dragged run (Eight of Hearts, Seven of Clubs) — not the
				// pile's top card (index 2).
				b, _ := s.tableauPiles[0].CardBoundsAt(1)
				return b.Min.X, b.Min.Y
			},
			upAt: func(s *SolitaireScene) (int, int) {
				b, _ := s.tableauPiles[1].TopBounds()
				return b.Min.X, b.Min.Y
			},
			wantBoard: func(t *testing.T, b *solitaire.Board) {
				t.Helper()
				if got := b.Tableau[0]; len(got.Cards) != 1 || got.FaceUp != 1 {
					t.Errorf("source pile = %+v; want 1 card, FaceUp=1 (the Nine flipped face-up)", got)
				}
				if got := b.Tableau[1]; len(got.Cards) != 3 {
					t.Errorf("expected the Eight-Seven run to move onto the Nine of Clubs, got %d cards", len(got.Cards))
				}
			},
		},
		{
			name: "dropping a multi-card run onto the foundation does nothing",
			board: &solitaire.Board{
				Tableau: [solitaire.TableauPileCount]solitaire.Pile{
					0: {Cards: []card.Card{
						{Suit: card.Club, Rank: card.Ace},
						{Suit: card.Heart, Rank: card.King}, // not really a legal run, but
					}, FaceUp: 2}, // Mediator shouldn't even need it to be legal to reject this drop.
				},
			},
			downAt: func(s *SolitaireScene) (int, int) {
				b, _ := s.tableauPiles[0].CardBoundsAt(0)
				return b.Min.X, b.Min.Y
			},
			upAt: func(s *SolitaireScene) (int, int) {
				b := s.foundationPile.Bounds()
				return b.Min.X, b.Min.Y
			},
			wantBoard: func(t *testing.T, b *solitaire.Board) {
				t.Helper()
				if len(b.Tableau[0].Cards) != 2 {
					t.Errorf("expected the multi-card run to not move to the foundation, got %d cards left", len(b.Tableau[0].Cards))
				}
				if len(b.Foundation[card.Club]) != 0 {
					t.Errorf("expected nothing to land on the foundation, got %d cards", len(b.Foundation[card.Club]))
				}
			},
		},
		{
			name: "dropping onto the same tableau pile does nothing",
			board: &solitaire.Board{
				Tableau: [solitaire.TableauPileCount]solitaire.Pile{
					0: {Cards: []card.Card{{Suit: card.Heart, Rank: card.Ten}}, FaceUp: 1},
				},
			},
			downAt: func(s *SolitaireScene) (int, int) {
				b, _ := s.tableauPiles[0].TopBounds()
				return b.Min.X, b.Min.Y
			},
			upAt: func(s *SolitaireScene) (int, int) {
				b := s.tableauPiles[0].Bounds()
				return b.Min.X, b.Min.Y
			},
			wantBoard: func(t *testing.T, b *solitaire.Board) {
				t.Helper()
				if len(b.Tableau[0].Cards) != 1 {
					t.Errorf("expected dropping a pile onto itself to be a no-op, got %d cards", len(b.Tableau[0].Cards))
				}
			},
		},
		{
			name:  "drop on empty space does nothing but still ends the drag",
			board: &solitaire.Board{Waste: []card.Card{{Suit: card.Spade, Rank: card.Ace}}},
			downAt: func(s *SolitaireScene) (int, int) {
				b, _ := s.wastePile.TopBounds()
				return b.Min.X, b.Min.Y
			},
			upAt: func(s *SolitaireScene) (int, int) {
				return -100, -100
			},
			wantBoard: func(t *testing.T, b *solitaire.Board) {
				t.Helper()
				if len(b.Waste) != 1 {
					t.Errorf("expected the drop to be ignored, got waste=%d", len(b.Waste))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestScene(t)
			s.setBoard(tt.board)
			s.syncComponents()

			dx, dy := tt.downAt(s)
			s.mediator.Handle(Event{Type: EventPointerDown, X: dx, Y: dy})
			if _, ok := s.mediator.Dragging(); !ok {
				t.Fatal("expected the pointer-down event to start a drag")
			}

			s.syncComponents()
			ux, uy := tt.upAt(s)
			s.mediator.Handle(Event{Type: EventPointerUp, X: ux, Y: uy})

			tt.wantBoard(t, s.board)

			if _, ok := s.mediator.Dragging(); ok {
				t.Error("expected the drag to have ended after the pointer-up event")
			}
		})
	}
}

// TestMediator_WiredAsRootEventHandler verifies the full chain end to
// end: NewSolitaireScene installs the Mediator as the RootComponent's
// event handler, so dispatching a real Event at a hit-tested leaf (going
// through Dispatch and RootComponent.HandleEvent, not calling any
// Mediator method directly) reaches it and acts on the board, exactly as
// SolitaireScene.Update does.
func TestMediator_WiredAsRootEventHandler(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{
		Stock: []card.Card{{Suit: card.Spade, Rank: card.Ace}},
	})
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

// TestSolitaireScene_Draw_WhenWon exercises Draw once Board.Won is true,
// so it also renders the win message.
func TestSolitaireScene_Draw_WhenWon(t *testing.T) {
	s := newTestScene(t)
	won := &solitaire.Board{}
	for _, suit := range []card.Suit{card.Spade, card.Heart, card.Diamond, card.Club} {
		for rank := card.Ace; rank <= card.King; rank++ {
			won.Foundation[suit] = append(won.Foundation[suit], card.Card{Suit: suit, Rank: rank})
		}
	}
	s.setBoard(won)
	if !s.board.Won() {
		t.Fatal("expected the constructed board to be won")
	}

	screen := ebiten.NewImage(1280, 720)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Draw panicked: %v", r)
		}
	}()

	s.Draw(screen)
}

// TestSolitaireScene_Draw_WithActiveDrag exercises Draw with a
// multi-card run being dragged, so it renders more than one floating
// card.
func TestSolitaireScene_Draw_WithActiveDrag(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{})
	s.board.Tableau[0] = solitaire.Pile{
		Cards: []card.Card{
			{Suit: card.Heart, Rank: card.Eight},
			{Suit: card.Club, Rank: card.Seven},
		},
		FaceUp: 2,
	}
	s.syncComponents()

	b, ok := s.tableauPiles[0].CardBoundsAt(0)
	if !ok {
		t.Fatal("expected tableau pile 0 to have a card at index 0")
	}
	s.mediator.Handle(Event{Type: EventPointerDown, X: b.Min.X, Y: b.Min.Y})
	dragState, ok := s.mediator.Dragging()
	if !ok {
		t.Fatal("expected a drag to be active")
	}
	if len(dragState.Cards) != 2 {
		t.Fatalf("expected the whole 2-card run to be dragged, got %d cards", len(dragState.Cards))
	}

	screen := ebiten.NewImage(1280, 720)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Draw panicked: %v", r)
		}
	}()

	s.Draw(screen)
}
