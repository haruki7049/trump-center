package game

import (
	"errors"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/haruki7049/trump-center/internal/scene"
)

type mockScene struct {
	updateCalled bool
	drawCalled   bool
	nextScene    scene.Scene
	errToReturn  error
}

func (m *mockScene) Update() (scene.Scene, error) {
	m.updateCalled = true
	return m.nextScene, m.errToReturn
}

func (m *mockScene) Draw(screen *ebiten.Image) {
	m.drawCalled = true
}

func TestConstants(t *testing.T) {
	if WINDOW_WIDTH != 1280 {
		t.Errorf("expected WINDOW_WIDTH to be 1280, got %d", WINDOW_WIDTH)
	}
	if WINDOW_HEIGHT != 720 {
		t.Errorf("expected WINDOW_HEIGHT to be 720, got %d", WINDOW_HEIGHT)
	}
	if WINDOW_TITLE != "Trump Center" {
		t.Errorf("expected WINDOW_TITLE to be 'Trump Center', got '%s'", WINDOW_TITLE)
	}
}

func TestNewGame(t *testing.T) {
	g, err := NewGame()
	if err != nil {
		t.Fatalf("expected no error from NewGame, got %v", err)
	}
	if g == nil {
		t.Fatalf("expected non-nil Game struct")
	}
	if g.scene == nil {
		t.Errorf("expected scene to be set in NewGame")
	}
}

func TestGame_Layout(t *testing.T) {
	g := &Game{}
	w, h := g.Layout(800, 600)
	if w != 800 || h != 600 {
		t.Errorf("expected Layout(800, 600) to return (800, 600), got (%d, %d)", w, h)
	}
}

func TestGame_Update(t *testing.T) {
	mock := &mockScene{}
	g := &Game{scene: mock}

	err := g.Update()
	if err != nil {
		t.Errorf("expected no error from Update, got %v", err)
	}
	if !mock.updateCalled {
		t.Errorf("expected mockScene.Update to be called")
	}

	// Test scene transition
	next := &mockScene{}
	g.scene = &mockScene{nextScene: next}
	err = g.Update()
	if err != nil {
		t.Errorf("expected no error during scene transition, got %v", err)
	}
	if g.scene != next {
		t.Errorf("expected scene to switch to next scene")
	}
}

func TestGame_Update_NilScene(t *testing.T) {
	g := &Game{scene: nil}
	err := g.Update()
	if err != nil {
		t.Fatalf("expected no error from Update with nil scene, got %v", err)
	}
	if g.scene == nil {
		t.Errorf("expected newGameScene to initialize scene when nil")
	}
}

func TestGame_Update_Error(t *testing.T) {
	expectedErr := errors.New("update error")
	mock := &mockScene{errToReturn: expectedErr}
	g := &Game{scene: mock}

	err := g.Update()
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestGame_Draw(t *testing.T) {
	screen := ebiten.NewImage(1280, 720)

	// Test Draw with non-nil scene
	mock := &mockScene{}
	g := &Game{scene: mock}
	g.Draw(screen)
	if !mock.drawCalled {
		t.Errorf("expected mockScene.Draw to be called")
	}

	// Test Draw with nil scene (should not panic)
	gNil := &Game{scene: nil}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Draw with nil scene panicked: %v", r)
		}
	}()
	gNil.Draw(screen)
}

func TestGame_newGameScene(t *testing.T) {
	g := &Game{}
	err := g.newGameScene()
	if err != nil {
		t.Fatalf("expected no error from newGameScene, got %v", err)
	}
	if g.scene == nil {
		t.Errorf("expected scene to be set by newGameScene")
	}
}
