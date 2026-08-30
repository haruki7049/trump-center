package game

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/haruki7049/trump-center/internal/scene"
	"github.com/haruki7049/trump-center/internal/scene/title"
)

const WINDOW_WIDTH = 1280
const WINDOW_HEIGHT = 720
const WINDOW_TITLE = "Trump Center"

// Run our trump game
func Run() {
	ebiten.SetWindowSize(WINDOW_WIDTH, WINDOW_HEIGHT)
	ebiten.SetWindowTitle(WINDOW_TITLE)

	game, err := NewGame()
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Game is the root ebiten.Game implementation. It only owns the currently
// active Scene and delegates Update/Draw to it.
type Game struct {
	scene scene.Scene
}

func NewGame() (*Game, error) {
	scene, err := title.NewTitleScene()
	if err != nil {
		return nil, err
	}

	result := &Game{scene: scene}
	return result, nil
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
