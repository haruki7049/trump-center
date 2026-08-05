package title

import (
	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/haruki7049/trump-center/internal/scene"
	"github.com/haruki7049/trump-center/internal/ui"
)

const titleMessage = "Trump Center"
const titleMessageSizeX = 7.5
const titleMessageSizeY = 7.5
const titleMessageTranslationX = 0
const titleMessageTranslationY = 0

type TitleScene struct {
	menuButton *ui.Button
	fontFace   *text.GoXFace
}

func NewTitleScene() *TitleScene {
	return &TitleScene{
		fontFace: text.NewGoXFace(bitmapfont.Face),
	}
}

func (s *TitleScene) Update() (scene.Scene, error) {
	return nil, nil
}

func (s *TitleScene) Draw(screen *ebiten.Image) {
	titleMessageOp := &text.DrawOptions{}
	titleMessageOp.GeoM.Scale(titleMessageSizeX, titleMessageSizeY)
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
