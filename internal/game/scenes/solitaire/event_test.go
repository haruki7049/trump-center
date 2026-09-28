package solitaire

import (
	"image"
	"testing"
)

func TestComponent_ParentLink(t *testing.T) {
	pile := NewPileComponent(image.Rect(0, 0, 10, 10))
	if got := pile.Parent(); got != nil {
		t.Errorf("Parent() before being added to a tree = %v; want nil", got)
	}

	root := NewRootComponent(pile)

	if got := pile.Parent(); got != Component(root) {
		t.Errorf("pile.Parent() = %v; want root", got)
	}
	if got := root.Parent(); got != nil {
		t.Errorf("root.Parent() = %v; want nil", got)
	}
}

func TestDispatch_BubblesToRoot(t *testing.T) {
	pile := NewPileComponent(image.Rect(0, 0, 10, 10))
	root := NewRootComponent(pile)

	var got Event
	called := false
	root.SetEventHandler(func(e Event) {
		called = true
		got = e
	})

	want := Event{Type: EventPointerDown, X: 5, Y: 5}
	if handled := Dispatch(pile, want); !handled {
		t.Fatal("expected the event fired at the leaf to be handled once it reaches Root")
	}
	if !called {
		t.Fatal("expected Root's event handler to be called")
	}
	if got != want {
		t.Errorf("handler received %+v; want %+v", got, want)
	}
}

func TestDispatch_NoHandlerInstalled(t *testing.T) {
	pile := NewPileComponent(image.Rect(0, 0, 10, 10))
	NewRootComponent(pile)

	if handled := Dispatch(pile, Event{Type: EventPointerUp}); handled {
		t.Error("expected Dispatch to report false when Root has no handler installed")
	}
}
