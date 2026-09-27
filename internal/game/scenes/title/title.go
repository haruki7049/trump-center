package title

import (
	"image/color"
	"log"

	"github.com/ebitenui/ebitenui"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/haruki7049/trump-center/assets"
	"github.com/haruki7049/trump-center/internal/scene"
)

type TitleScene struct {
	fontFace text.Face
	ui       *ebitenui.UI
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
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(16),
		)),
	)
	eui := &ebitenui.UI{Container: rootContainer}

	helloWorldlabel := widget.NewText(widget.TextOpts.Text("Hello, world!", &ts.fontFace, color.White))
	rootContainer.AddChild(helloWorldlabel)

	startButton := widget.NewButton(
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    eimage.NewNineSliceColor(color.NRGBA{0x40, 0x40, 0x40, 0xff}),
			Hover:   eimage.NewNineSliceColor(color.NRGBA{0x60, 0x60, 0x60, 0xff}),
			Pressed: eimage.NewNineSliceColor(color.NRGBA{0x20, 0x20, 0x20, 0xff}),
		}),
		widget.ButtonOpts.Text("Start", &ts.fontFace, &widget.ButtonTextColor{Idle: color.White}),
		widget.ButtonOpts.TextPadding(&widget.Insets{Left: 16, Right: 16, Top: 8, Bottom: 8}),
		widget.ButtonOpts.ClickedHandler(func(_ *widget.ButtonClickedEventArgs) {
			log.Println("start button clicked")
		}),
	)
	rootContainer.AddChild(startButton)

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
