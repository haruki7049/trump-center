package scene_test

import (
	"errors"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/haruki7049/trump-center/internal/scene"
)

type dummyScene struct {
	updateCalled bool
	drawCalled   bool
	nextScene    scene.Scene
	errToReturn  error
}

func (s *dummyScene) Update() (scene.Scene, error) {
	s.updateCalled = true
	return s.nextScene, s.errToReturn
}

func (s *dummyScene) Draw(screen *ebiten.Image) {
	s.drawCalled = true
}

func TestSceneInterface(t *testing.T) {
	var s scene.Scene = &dummyScene{}

	// Test Update
	next, err := s.Update()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if next != nil {
		t.Errorf("expected nil next scene, got %v", next)
	}

	dummy, ok := s.(*dummyScene)
	if !ok || !dummy.updateCalled {
		t.Errorf("expected Update to have been called")
	}

	// Test Draw
	screen := ebiten.NewImage(100, 100)
	s.Draw(screen)
	if !dummy.drawCalled {
		t.Errorf("expected Draw to have been called")
	}
}

func TestSceneInterface_WithError(t *testing.T) {
	expectedErr := errors.New("scene error")
	dummy := &dummyScene{errToReturn: expectedErr}
	var s scene.Scene = dummy

	_, err := s.Update()
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}
