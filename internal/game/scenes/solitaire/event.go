package solitaire

// EventType identifies what kind of thing happened to a component.
type EventType int

const (
	// EventPointerDown fires when the player presses the mouse button
	// while the pointer is over a component.
	EventPointerDown EventType = iota
	// EventPointerUp fires when the player releases the mouse button
	// while the pointer is over a component.
	EventPointerUp
)

// Event describes something that happened to a component, to be bubbled
// up the component tree via Chain of Responsibility (see Dispatch) until
// something handles it.
type Event struct {
	Type EventType
	X, Y int
}

// Dispatch bubbles e up the Chain of Responsibility, starting at source
// and walking Parent() links until some component's HandleEvent reports
// that it handled e, or the chain is exhausted. It reports whether e was
// handled.
func Dispatch(source Component, e Event) bool {
	for c := source; c != nil; c = c.Parent() {
		if c.HandleEvent(e) {
			return true
		}
	}
	return false
}
