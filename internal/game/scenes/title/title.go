package title

import (
	"image/color"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/haruki7049/trump-center/assets"
	"github.com/haruki7049/trump-center/internal/scene"
	"github.com/haruki7049/trump-center/internal/ui"
)

type TitleScene struct {
	menuButton *ui.Button
	fontFace   text.Face
	ui         *ebitenui.UI
}

// Creates the new TitleScene value with initial member variables
func NewTitleScene() (*TitleScene, error) {
	var ts TitleScene
	if err := ts.newTitleSceneFontFace(); err != nil {
		return nil, err
	}

	ts.newTitleSceneUi()

	return &ts, nil
}

func (ts *TitleScene) newTitleSceneUi() {
	rootContainer := widget.NewContainer()
	eui := &ebitenui.UI{Container: rootContainer}

	helloWorldlabel := widget.NewText(widget.TextOpts.Text("Hello, world!", &ts.fontFace, color.White))
	rootContainer.AddChild(helloWorldlabel)
	ts.ui = eui
}

func (ts *TitleScene) newTitleSceneFontFace() error {
	fontFace, err := loadFont("fonts/DotGothic16/DotGothic16-Regular.ttf")
	if err != nil {
		return err
	}

	ts.fontFace = fontFace
	return nil
}

func loadFont(filename string) (*text.GoTextFace, error) {
	// Read ttf file
	f, err := assets.Assets.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	src, err := text.NewGoTextFaceSource(f)
	if err != nil {
		return nil, err
	}

	fontFace := text.GoTextFace{Source: src, Size: 30}
	return &fontFace, nil
}

func (s *TitleScene) Update() (scene.Scene, error) {
	s.ui.Update()
	return nil, nil
}

func (s *TitleScene) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
}

// lineSpacing returns the standard line spacing for fontFace, used by
// every scene when drawing text.
func (s *TitleScene) lineSpacing() float64 {
	m := s.fontFace.Metrics()
	return m.HLineGap + m.HAscent + m.HDescent
}
