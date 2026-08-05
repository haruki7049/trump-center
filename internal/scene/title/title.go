package title

import (
	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/haruki7049/trump-center/internal/scene"
	"github.com/haruki7049/trump-center/internal/ui"
)

const titleMessage = "Trump Center"

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
	titleMessageOp.GeoM.Translate(20, 20)
	titleMessageOp.LineSpacing = s.lineSpacing()
	text.Draw(screen, titleMessage, s.fontFace, titleMessageOp)

	ebitenutil.DebugPrint(screen, "Hello, world from TitleScene!!")
}

// lineSpacing returns the standard line spacing for fontFace, used by
// every scene when drawing text.
func (s *TitleScene) lineSpacing() float64 {
	m := s.fontFace.Metrics()
	return m.HLineGap + m.HAscent + m.HDescent
}
