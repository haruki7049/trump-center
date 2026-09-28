package solitaire

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestPileComponent_Draw(t *testing.T) {
	back, err := loadImage("cards/back.png")
	if err != nil {
		t.Fatalf("failed to load test image: %v", err)
	}

	p := NewPileComponent(image.Rect(0, 0, cardWidth, cardHeight))
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

func TestRootComponent_HitTest(t *testing.T) {
	p1 := NewPileComponent(image.Rect(0, 0, 100, 100))
	p2 := NewPileComponent(image.Rect(200, 200, 300, 300))
	root := NewRootComponent(p1, p2)

	if got := root.HitTest(50, 50); got != Component(p1) {
		t.Errorf("HitTest(50, 50) = %v; want p1", got)
	}
	if got := root.HitTest(250, 250); got != Component(p2) {
		t.Errorf("HitTest(250, 250) = %v; want p2", got)
	}
	if got := root.HitTest(150, 150); got != nil {
		t.Errorf("HitTest(150, 150) = %v; want nil", got)
	}
}

func TestPileComponent_TopBounds(t *testing.T) {
	p := NewPileComponent(image.Rect(0, 0, cardWidth, cardHeight))

	if _, ok := p.TopBounds(); ok {
		t.Error("expected TopBounds to report false for an empty pile")
	}

	p.SetCards([]CardDraw{
		{Image: nil, X: 0, Y: 0},
		{Image: nil, X: 10, Y: 20},
	})

	got, ok := p.TopBounds()
	if !ok {
		t.Fatal("expected TopBounds to report true once cards are set")
	}
	want := image.Rect(10, 20, 10+cardWidth, 20+cardHeight)
	if got != want {
		t.Errorf("TopBounds() = %v; want %v", got, want)
	}
}

func TestRootComponent_DrawDelegatesToChildren(t *testing.T) {
	back, err := loadImage("cards/back.png")
	if err != nil {
		t.Fatalf("failed to load test image: %v", err)
	}

	p1 := NewPileComponent(image.Rect(0, 0, cardWidth, cardHeight))
	p1.SetCards([]CardDraw{{Image: back, X: 0, Y: 0}})
	p2 := NewPileComponent(image.Rect(50, 50, 50+cardWidth, 50+cardHeight))
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
