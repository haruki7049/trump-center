package ui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	buttonColor      = color.RGBA{R: 0x40, G: 0x40, B: 0x40, A: 0xff}
	buttonHoverColor = color.RGBA{R: 0x60, G: 0x60, B: 0x60, A: 0xff}
)

// Button is a simple clickable rectangular UI element with a text label.
type Button struct {
	Bounds image.Rectangle
	Label  string
}

// Contains reports whether the given point is inside the button.
func (b *Button) Contains(x, y int) bool {
	return image.Pt(x, y).In(b.Bounds)
}

// Clicked reports whether the button was clicked during the current frame.
func (b *Button) Clicked() bool {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return false
	}
	x, y := ebiten.CursorPosition()
	return b.Contains(x, y)
}

// Draw renders the button as a filled rectangle with a label inside it.
func (b *Button) Draw(screen *ebiten.Image, face text.Face) {
	x, y := ebiten.CursorPosition()
	clr := buttonColor
	if b.Contains(x, y) {
		clr = buttonHoverColor
	}

	vector.FillRect(
		screen,
		float32(b.Bounds.Min.X),
		float32(b.Bounds.Min.Y),
		float32(b.Bounds.Dx()),
		float32(b.Bounds.Dy()),
		clr,
		false,
	)

	op := &text.DrawOptions{}
	op.GeoM.Translate(
		float64(b.Bounds.Min.X)+8,
		float64(b.Bounds.Min.Y)+8,
	)
	text.Draw(screen, b.Label, face, op)
}
