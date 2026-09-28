package solitaire

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewSolitaireScene_BuildsUI(t *testing.T) {
	s := newTestScene(t)
	if s.ui == nil {
		t.Error("expected ui to be initialized")
	}
	if s.labelFontFace == nil {
		t.Error("expected labelFontFace to be initialized")
	}
}

// TestSolitaireScene_Draw_DrawsLabels exercises Draw's label-drawing
// path directly, since it isn't otherwise covered by any behavior
// assertion (only that it doesn't panic).
func TestSolitaireScene_Draw_DrawsLabels(t *testing.T) {
	s := newTestScene(t)
	screen := ebiten.NewImage(1280, 720)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("drawLabels panicked: %v", r)
		}
	}()

	s.drawLabels(screen)
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
