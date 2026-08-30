package title

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/haruki7049/trump-center/assets"
	"github.com/haruki7049/trump-center/internal/scene"
	"github.com/haruki7049/trump-center/internal/ui"
)

const titleMessage = "Trump Center"
const titleMessageTranslationX = 0
const titleMessageTranslationY = 0

type TitleScene struct {
	menuButton *ui.Button
	fontFace   *text.GoTextFace
}

// Creates the new TitleScene value with initial member variables
func NewTitleScene() (*TitleScene, error) {
	fontFace, err := loadFont("fonts/DotGothic16/DotGothic16-Regular.ttf")
	if err != nil {
		return nil, err
	}

	result := &TitleScene{fontFace: fontFace}
	return result, nil
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
	return nil, nil
}

func (s *TitleScene) Draw(screen *ebiten.Image) {
	titleMessageOp := &text.DrawOptions{}
	titleMessageOp.GeoM.Translate(titleMessageTranslationX, titleMessageTranslationY)
	titleMessageOp.LineSpacing = s.lineSpacing()
	text.Draw(screen, titleMessage, s.fontFace, titleMessageOp)
}

// lineSpacing returns the standard line spacing for fontFace, used by
// every scene when drawing text.
func (s *TitleScene) lineSpacing() float64 {
	m := s.fontFace.Metrics()
	return m.HLineGap + m.HAscent + m.HDescent
}
