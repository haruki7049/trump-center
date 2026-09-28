package solitaire

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Component is a node in the board's Composite view hierarchy. Every
// component knows how to draw itself, report the screen area it
// occupies, and, if it is a container, how to reach its children; it
// holds no game-rule logic (see PileComponent).
type Component interface {
	Draw(screen *ebiten.Image)
	Children() []Component
	Bounds() image.Rectangle
}

// RootComponent is the root of the board's component tree. It has no
// visual representation of its own; it only draws its children in order
// and hit-tests them.
type RootComponent struct {
	children []Component
}

// NewRootComponent builds a RootComponent with the given children.
func NewRootComponent(children ...Component) *RootComponent {
	return &RootComponent{children: children}
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
}

// NewPileComponent builds an empty PileComponent occupying bounds.
func NewPileComponent(bounds image.Rectangle) *PileComponent {
	return &PileComponent{bounds: bounds}
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
