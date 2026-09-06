package title

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewTitleScene(t *testing.T) {
	ts, err := NewTitleScene()
	if err != nil {
		t.Fatalf("expected no error when creating TitleScene, got %v", err)
	}
	if ts == nil {
		t.Fatalf("expected non-nil TitleScene")
	}
	if ts.fontFace == nil {
		t.Errorf("expected fontFace to be initialized")
	}
	if ts.ui == nil {
		t.Errorf("expected ui to be initialized")
	}
}

func TestTitleScene_Update(t *testing.T) {
	ts, err := NewTitleScene()
	if err != nil {
		t.Fatalf("failed to create TitleScene: %v", err)
	}

	nextScene, err := ts.Update()
	if err != nil {
		t.Errorf("expected no error during Update, got %v", err)
	}
	if nextScene != nil {
		t.Errorf("expected nil next scene from Update, got %v", nextScene)
	}
}

func TestTitleScene_Draw(t *testing.T) {
	ts, err := NewTitleScene()
	if err != nil {
		t.Fatalf("failed to create TitleScene: %v", err)
	}

	screen := ebiten.NewImage(1280, 720)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Draw panicked: %v", r)
		}
	}()

	ts.Draw(screen)
}

func TestLoadFont_Success(t *testing.T) {
	face, err := loadFont("fonts/DotGothic16/DotGothic16-Regular.ttf")
	if err != nil {
		t.Fatalf("expected successful font loading, got error: %v", err)
	}
	if face == nil {
		t.Fatalf("expected non-nil GoTextFace")
	}
	if face.Size != 30 {
		t.Errorf("expected font size 30, got %f", face.Size)
	}
}

func TestLoadFont_NotFound(t *testing.T) {
	_, err := loadFont("nonexistent/font.ttf")
	if err == nil {
		t.Errorf("expected error when loading non-existent font, got nil")
	}
}
