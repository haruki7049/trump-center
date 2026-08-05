package scene

import "github.com/hajimehoshi/ebiten/v2"

// Scene represents a single screen of the game (title, menu, etc).
// Update returns the next Scene to switch to, or nil to stay on the
// current one.
type Scene interface {
	Update() (Scene, error)
	Draw(screen *ebiten.Image)
}
