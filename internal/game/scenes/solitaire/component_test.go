package solitaire

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestPileComponent_Draw(t *testing.T) {
	back, err := loadImage("cards/back.png")
	if err != nil {
		t.Fatalf("failed to load test image: %v", err)
	}

	p := NewPileComponent()
	p.SetCards([]CardDraw{{Image: back, X: 10, Y: 20}})

	if got := p.Children(); got != nil {
		t.Errorf("expected PileComponent to have no children, got %v", got)
	}

	screen := ebiten.NewImage(100, 100)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Draw panicked: %v", r)
		}
	}()

	p.Draw(screen)
}

func TestRootComponent_DrawDelegatesToChildren(t *testing.T) {
	back, err := loadImage("cards/back.png")
	if err != nil {
		t.Fatalf("failed to load test image: %v", err)
	}

	p1 := NewPileComponent()
	p1.SetCards([]CardDraw{{Image: back, X: 0, Y: 0}})
	p2 := NewPileComponent()
	p2.SetCards([]CardDraw{{Image: back, X: 50, Y: 50}})

	root := NewRootComponent(p1, p2)
	if got := len(root.Children()); got != 2 {
		t.Fatalf("expected 2 children, got %d", got)
	}

	screen := ebiten.NewImage(100, 100)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Draw panicked: %v", r)
		}
	}()

	root.Draw(screen)
}
