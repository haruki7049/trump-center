package solitaire

import "testing"

func TestNewSolitaireScene_BuildsUI(t *testing.T) {
	s := newTestScene(t)
	if s.ui == nil {
		t.Error("expected ui to be initialized")
	}
}

// TestSolitaireScene_NewGameRequested_Redeals verifies that setting
// newGameRequested (as the "New Game" button's click handler does) and
// calling Update deals a fresh board and rebuilds the component tree and
// Mediator around it, the same way NewSolitaireScene does at startup.
func TestSolitaireScene_NewGameRequested_Redeals(t *testing.T) {
	s := newTestScene(t)
	oldBoard := s.board
	oldRoot := s.root
	oldMediator := s.mediator

	s.newGameRequested = true
	if _, err := s.Update(); err != nil {
		t.Fatalf("unexpected error from Update: %v", err)
	}

	if s.newGameRequested {
		t.Error("expected newGameRequested to be cleared after Update")
	}
	if s.board == oldBoard {
		t.Error("expected a fresh Board to be dealt")
	}
	if s.root == oldRoot {
		t.Error("expected the component tree to be rebuilt")
	}
	if s.mediator == oldMediator {
		t.Error("expected the Mediator to be rebuilt")
	}
}
