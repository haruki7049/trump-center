package solitaire

import (
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

// Component is a node in the board's Composite view hierarchy. Every
// component knows how to draw itself, report the screen area it
// occupies, and, if it is a container, how to reach its children; it
// holds no game-rule logic (see PileComponent). Parent and HandleEvent
// support bubbling Events up the tree via Chain of Responsibility (see
// Dispatch in event.go): a component that does not handle an event
// simply declines it (HandleEvent returns false), leaving Dispatch to
// retry at Parent().
type Component interface {
	Draw(screen *ebiten.Image)
	Children() []Component
	Bounds() image.Rectangle
	Parent() Component
	HandleEvent(e Event) bool
}

// parentSetter is implemented by concrete components so that a container
// (currently only RootComponent) can link itself as their parent when
// they are added to the tree. It is unexported because the parent link
// reflects tree structure and must only be set by the tree itself, never
// by external callers.
type parentSetter interface {
	setParent(p Component)
}

// RootComponent is the root of the board's component tree. It has no
// visual representation of its own; it only draws its children in order,
// hit-tests them, and is the terminal link in the event-bubbling Chain of
// Responsibility.
type RootComponent struct {
	children []Component
	onEvent  func(Event)
}

// NewRootComponent builds a RootComponent with the given children, and
// links each of them to it via Parent().
func NewRootComponent(children ...Component) *RootComponent {
	r := &RootComponent{children: children}
	for _, c := range children {
		if ps, ok := c.(parentSetter); ok {
			ps.setParent(r)
		}
	}
	return r
}

// Parent always returns nil: RootComponent is the top of the tree.
func (r *RootComponent) Parent() Component {
	return nil
}

// SetEventHandler installs the terminal handler for events that bubble
// all the way up to Root without being handled along the way. From Step 4
// of the architecture refactor onward this is the Mediator; until then,
// HandleEvent just logs unhandled events when no handler is installed.
func (r *RootComponent) SetEventHandler(h func(Event)) {
	r.onEvent = h
}

// HandleEvent is the terminal link in the Chain of Responsibility: if a
// handler has been installed via SetEventHandler, it delegates to it and
// reports the event as handled; otherwise it logs the event as unhandled
// and reports it as not handled.
func (r *RootComponent) HandleEvent(e Event) bool {
	if r.onEvent != nil {
		r.onEvent(e)
		return true
	}
	log.Printf("solitaire: unhandled event reached Root: %+v", e)
	return false
}

func (r *RootComponent) Children() []Component {
	return r.children
}

// Bounds is unused for RootComponent: HitTest checks its children, not
// Root itself.
func (r *RootComponent) Bounds() image.Rectangle {
	return image.Rectangle{}
}

func (r *RootComponent) Draw(screen *ebiten.Image) {
	for _, c := range r.children {
		c.Draw(screen)
	}
}

// HitTest returns the first child whose Bounds contains (x, y), or nil
// if none does.
func (r *RootComponent) HitTest(x, y int) Component {
	pt := image.Pt(x, y)
	for _, c := range r.children {
		if pt.In(c.Bounds()) {
			return c
		}
	}
	return nil
}

// CardDraw is a fully-resolved instruction to draw one card image at a
// specific screen position.
type CardDraw struct {
	Image *ebiten.Image
	X, Y  int
}

// PileComponent is a Passive View (MVP): it only stores and renders the
// CardDraw values it was last given via SetCards, within a fixed screen
// area set at construction. It has no knowledge of game rules, piles, or
// the board — those live in internal/solitaire and (from Step 4 onward)
// the Mediator.
type PileComponent struct {
	bounds image.Rectangle
	cards  []CardDraw
	parent Component
}

// NewPileComponent builds an empty PileComponent occupying bounds. Its
// Parent is nil until it is passed to NewRootComponent.
func NewPileComponent(bounds image.Rectangle) *PileComponent {
	return &PileComponent{bounds: bounds}
}

func (p *PileComponent) setParent(parent Component) {
	p.parent = parent
}

func (p *PileComponent) Parent() Component {
	return p.parent
}

// HandleEvent never handles anything: PileComponent is a Passive View
// with no logic, so every event bubbles past it toward Root.
func (p *PileComponent) HandleEvent(e Event) bool {
	return false
}

// SetCards replaces the cards this component draws.
func (p *PileComponent) SetCards(cards []CardDraw) {
	p.cards = cards
}

func (p *PileComponent) Children() []Component {
	return nil
}

func (p *PileComponent) Bounds() image.Rectangle {
	return p.bounds
}

// TopBounds returns the screen area of the last (topmost) card currently
// drawn by this pile, if any.
func (p *PileComponent) TopBounds() (image.Rectangle, bool) {
	if len(p.cards) == 0 {
		return image.Rectangle{}, false
	}

	top := p.cards[len(p.cards)-1]
	return image.Rect(top.X, top.Y, top.X+cardWidth, top.Y+cardHeight), true
}

func (p *PileComponent) Draw(screen *ebiten.Image) {
	for _, c := range p.cards {
		drawCard(screen, c.Image, c.X, c.Y)
	}
}
