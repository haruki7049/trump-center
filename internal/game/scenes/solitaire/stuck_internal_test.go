package solitaire

import (
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/haruki7049/trump-center/internal/card"
	"github.com/haruki7049/trump-center/internal/solitaire"
)

// stuckBoard returns a small stuck position: the Two of Hearts can't go
// anywhere, so the Ace of Spades beneath it never flips; the Five of
// Diamonds in the stock can't go anywhere either; and tableau piles 1-6
// each hold a lone King that can't move or take anything.
func stuckBoard() *solitaire.Board {
	b := &solitaire.Board{
		Stock: []card.Card{{Suit: card.Diamond, Rank: card.Five}},
	}
	b.Tableau[0] = solitaire.Pile{Cards: []card.Card{
		{Suit: card.Spade, Rank: card.Ace},
		{Suit: card.Heart, Rank: card.Two},
	}, FaceUp: 1}
	for i := 1; i < solitaire.TableauPileCount; i++ {
		b.Tableau[i] = solitaire.Pile{Cards: []card.Card{{Suit: card.Club, Rank: card.King}}, FaceUp: 1}
	}
	return b
}

func TestMediator_Stuck(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(stuckBoard())
	s.mediator.waitStuck()

	if !s.mediator.Stuck() {
		t.Error("expected a stuck board to be reported stuck")
	}
}

func TestMediator_Stuck_NotStuck(t *testing.T) {
	s := newTestScene(t)
	b := stuckBoard()
	b.Stock = append(b.Stock, card.Card{Suit: card.Diamond, Rank: card.Ace})
	s.setBoard(b)
	s.mediator.waitStuck()

	if s.mediator.Stuck() {
		t.Error("expected a board with a playable Ace in the stock to not be stuck")
	}
}

func TestMediator_Stuck_PickedUpByTick(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(stuckBoard())

	// Tick polls without blocking, so keep ticking until the background
	// check has delivered its result.
	for s.mediator.stuckResult != nil {
		s.mediator.Tick()
	}
	if !s.mediator.Stuck() {
		t.Error("expected Tick to pick up the background result")
	}
}

func TestMediator_Stuck_RecheckedAfterMove(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(stuckBoard())
	s.mediator.waitStuck()
	if !s.mediator.Stuck() {
		t.Fatal("expected the starting board to be stuck")
	}

	// Drawing from the stock is a change to the board, so it must restart
	// the check, clearing the stale "stuck" until the new result arrives.
	s.syncComponents()
	pt := s.stockPile.Bounds().Min
	s.mediator.Handle(Event{Type: EventPointerDown, X: pt.X, Y: pt.Y})
	if s.mediator.Stuck() {
		t.Error("expected Stuck to be cleared while the board is rechecked")
	}
	s.mediator.waitStuck()
	if !s.mediator.Stuck() {
		t.Error("expected the board to still be stuck after drawing the unplayable card")
	}
}

func TestMediator_Stuck_RecheckedAfterSuccessfulDrop(t *testing.T) {
	s := newTestScene(t)
	b := stuckBoard()
	// A red Queen in the waste can go onto any of the black Kings.
	b.Waste = []card.Card{{Suit: card.Heart, Rank: card.Queen}}
	s.setBoard(b)
	s.mediator.waitStuck()
	s.syncComponents()

	top, _ := s.wastePile.TopBounds()
	s.mediator.Handle(Event{Type: EventPointerDown, X: top.Min.X, Y: top.Min.Y})
	dropAt, _ := s.tableauPiles[1].TopBounds()
	s.mediator.Handle(Event{Type: EventPointerUp, X: dropAt.Min.X + 1, Y: dropAt.Min.Y + 1})

	if len(s.board.Tableau[1].Cards) != 2 {
		t.Fatalf("expected the Queen to be moved, got pile 1 = %+v", s.board.Tableau[1])
	}
	if s.mediator.stuckResult == nil {
		t.Error("expected a successful drop to restart the Stuck check")
	}
}

func TestSolitaireScene_Draw_WhenStuck(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(stuckBoard())
	s.mediator.waitStuck()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Draw panicked while stuck: %v", r)
		}
	}()
	s.Draw(ebiten.NewImage(1280, 720))
}

func TestSolitaireScene_BoardMessage(t *testing.T) {
	won := &solitaire.Board{}
	for suit := range won.Foundation {
		for r := card.Ace; r <= card.King; r++ {
			won.Foundation[suit] = append(won.Foundation[suit], card.Card{Suit: card.Suit(suit), Rank: r})
		}
	}
	playable := stuckBoard()
	playable.Stock = append(playable.Stock, card.Card{Suit: card.Diamond, Rank: card.Ace})

	tests := []struct {
		name  string
		board *solitaire.Board
		want  boardMessage
	}{
		{"won", won, boardMessageWon},
		{"stuck", stuckBoard(), boardMessageStuck},
		{"moves remain", playable, boardMessageNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestScene(t)
			s.setBoard(tt.board)
			s.mediator.waitStuck()

			if got := s.boardMessage(); got != tt.want {
				t.Errorf("boardMessage() = %v; want %v", got, tt.want)
			}
		})
	}
}

// TestSolitaireScene_StuckMessageAppearsAfterLastMove plays the move that
// leaves the board stuck, then runs the scene's own Update loop (as the
// game would) and checks the message appears within a second, and that
// it wasn't shown before the move while a move remained.
func TestSolitaireScene_StuckMessageAppearsAfterLastMove(t *testing.T) {
	s := newTestScene(t)
	b := stuckBoard()
	b.Waste = []card.Card{{Suit: card.Diamond, Rank: card.Ace}}
	s.setBoard(b)
	s.mediator.waitStuck()
	s.syncComponents()

	if got := s.boardMessage(); got != boardMessageNone {
		t.Fatalf("boardMessage() = %v before the last move; want none", got)
	}

	top, _ := s.wastePile.TopBounds()
	s.mediator.Handle(Event{Type: EventPointerDown, X: top.Min.X, Y: top.Min.Y})
	s.mediator.Handle(Event{Type: EventPointerUp, X: foundationOriginX + 1, Y: foundationOriginY + 1})
	if len(s.board.Foundation[card.Diamond]) != 1 {
		t.Fatal("expected the Ace to be moved to the foundation")
	}

	deadline := time.Now().Add(time.Second)
	for s.boardMessage() != boardMessageStuck {
		if time.Now().After(deadline) {
			t.Fatal("expected the stuck message within a second of the last move")
		}
		if _, err := s.Update(); err != nil {
			t.Fatalf("unexpected error from Update: %v", err)
		}
	}
}
