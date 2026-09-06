package ui_test

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/haruki7049/trump-center/assets"
	"github.com/haruki7049/trump-center/internal/ui"
)

func TestButton_Contains(t *testing.T) {
	btn := &ui.Button{
		Bounds: image.Rect(10, 20, 100, 50),
		Label:  "Test",
	}

	tests := []struct {
		name string
		x    int
		y    int
		want bool
	}{
		{"Inside center", 50, 30, true},
		{"Top-left inside corner", 10, 20, true},
		{"Bottom-right outside border", 100, 50, false},
		{"Far left", 5, 30, false},
		{"Far right", 105, 30, false},
		{"Above", 50, 10, false},
		{"Below", 50, 60, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := btn.Contains(tt.x, tt.y)
			if got != tt.want {
				t.Errorf("Contains(%d, %d) = %v; want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestButton_Clicked(t *testing.T) {
	btn := &ui.Button{
		Bounds: image.Rect(10, 20, 100, 50),
		Label:  "Test",
	}

	// Without mouse input simulation, Clicked should return false
	if btn.Clicked() {
		t.Errorf("expected Clicked() to be false when mouse is not pressed")
	}
}

func TestButton_Draw(t *testing.T) {
	btn := &ui.Button{
		Bounds: image.Rect(10, 20, 100, 50),
		Label:  "Test Button",
	}

	f, err := assets.Assets.Open("fonts/DotGothic16/DotGothic16-Regular.ttf")
	if err != nil {
		t.Fatalf("failed to open font asset: %v", err)
	}
	defer f.Close()

	src, err := text.NewGoTextFaceSource(f)
	if err != nil {
		t.Fatalf("failed to create text face source: %v", err)
	}

	face := &text.GoTextFace{Source: src, Size: 16}
	screen := ebiten.NewImage(200, 200)

	// Test Draw execution without panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Draw panicked: %v", r)
		}
	}()

	btn.Draw(screen, face)
}

func TestButton_Draw_Hover(t *testing.T) {
	// Bounds including (0,0) which is default Ebitengine cursor position in tests
	btn := &ui.Button{
		Bounds: image.Rect(-10, -10, 50, 50),
		Label:  "Hover Button",
	}

	f, err := assets.Assets.Open("fonts/DotGothic16/DotGothic16-Regular.ttf")
	if err != nil {
		t.Fatalf("failed to open font asset: %v", err)
	}
	defer f.Close()

	src, err := text.NewGoTextFaceSource(f)
	if err != nil {
		t.Fatalf("failed to create text face source: %v", err)
	}

	face := &text.GoTextFace{Source: src, Size: 16}
	screen := ebiten.NewImage(200, 200)

	btn.Draw(screen, face)
}
