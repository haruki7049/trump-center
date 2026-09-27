package solitaire_test

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/haruki7049/trump-center/internal/game/scenes/solitaire"
)

func TestNewSolitaireScene(t *testing.T) {
	s, err := solitaire.NewSolitaireScene()
	if err != nil {
		t.Fatalf("expected no error when creating SolitaireScene, got %v", err)
	}
	if s == nil {
		t.Fatalf("expected non-nil SolitaireScene")
	}
}

func TestSolitaireScene_Update(t *testing.T) {
	s, err := solitaire.NewSolitaireScene()
	if err != nil {
		t.Fatalf("failed to create SolitaireScene: %v", err)
	}

	next, err := s.Update()
	if err != nil {
		t.Errorf("expected no error during Update, got %v", err)
	}
	if next != nil {
		t.Errorf("expected nil next scene from Update, got %v", next)
	}
}

func TestSolitaireScene_Draw(t *testing.T) {
	s, err := solitaire.NewSolitaireScene()
	if err != nil {
		t.Fatalf("failed to create SolitaireScene: %v", err)
	}

	screen := ebiten.NewImage(1280, 720)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Draw panicked: %v", r)
		}
	}()

	s.Draw(screen)
}
