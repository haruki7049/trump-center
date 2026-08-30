package game

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/haruki7049/trump-center/internal/game/scenes/title"
	"github.com/haruki7049/trump-center/internal/scene"
)

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

func (g *Game) newGameScene() error {
	scene, err := title.NewTitleScene()
	if err != nil {
		return err
	}

	g.scene = scene
	return nil
}

func (g *Game) Update() error {
	if g.scene == nil {
		if err := g.newGameScene(); err != nil {
			return err
		}
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
