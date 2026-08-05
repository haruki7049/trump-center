package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/haruki7049/trump-center/internal/scene"
	"github.com/haruki7049/trump-center/internal/scene/title"
)

const WINDOW_WIDTH = 1280
const WINDOW_HEIGHT = 720
const WINDOW_TITLE = "Trump Center"

// Run our trump game
func Run() error {
	ebiten.SetWindowSize(WINDOW_WIDTH, WINDOW_HEIGHT)
	ebiten.SetWindowTitle(WINDOW_TITLE)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		return err
	}

	return nil
}

// Game is the root ebiten.Game implementation. It only owns the currently
// active Scene and delegates Update/Draw to it.
type Game struct {
	scene scene.Scene
}

func NewGame() *Game {
	return &Game{
		scene: title.NewTitleScene(),
	}
}

type NullPointerOnSceneError struct{}

func (e NullPointerOnSceneError) Error() string {
	return "g.scene var is null"
}

func (g *Game) Update() error {
	if g.scene == nil {
		return NullPointerOnSceneError{}
	}

	next, err := g.scene.Update()
	if err != nil {
		return err
	}
	if next != nil {
		g.scene = next
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.scene == nil {
		return
	}

	g.scene.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}
