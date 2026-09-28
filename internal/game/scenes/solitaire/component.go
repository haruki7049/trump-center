package solitaire

import "github.com/hajimehoshi/ebiten/v2"

// Component is a node in the board's Composite view hierarchy. Every
// component knows how to draw itself and, if it is a container, how to
// reach its children; it holds no game-rule logic (see PileComponent).
type Component interface {
	Draw(screen *ebiten.Image)
	Children() []Component
}

// RootComponent is the root of the board's component tree. It has no
// visual representation of its own; it only draws its children in order.
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

func (r *RootComponent) Draw(screen *ebiten.Image) {
	for _, c := range r.children {
		c.Draw(screen)
	}
}

// CardDraw is a fully-resolved instruction to draw one card image at a
// specific screen position.
type CardDraw struct {
	Image *ebiten.Image
	X, Y  int
}

// PileComponent is a Passive View (MVP): it only stores and renders the
// CardDraw values it was last given via SetCards. It has no knowledge of
// game rules, piles, or the board — those live in internal/solitaire and
// (from Step 4 onward) the Mediator.
type PileComponent struct {
	cards []CardDraw
}

// NewPileComponent builds an empty PileComponent.
func NewPileComponent() *PileComponent {
	return &PileComponent{}
}

// SetCards replaces the cards this component draws.
func (p *PileComponent) SetCards(cards []CardDraw) {
	p.cards = cards
}

func (p *PileComponent) Children() []Component {
	return nil
}

func (p *PileComponent) Draw(screen *ebiten.Image) {
	for _, c := range p.cards {
		drawCard(screen, c.Image, c.X, c.Y)
	}
}
