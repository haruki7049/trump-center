// Package solitaire is the scene that draws a Klondike solitaire board.
// Cards can be drawn from the stock pile by clicking it; dragging cards
// between piles is not implemented yet.
package solitaire

import (
	"image"
	_ "image/png"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/haruki7049/trump-center/assets"
	"github.com/haruki7049/trump-center/internal/card"
	"github.com/haruki7049/trump-center/internal/scene"
	"github.com/haruki7049/trump-center/internal/solitaire"
)

const (
	// cardScale shrinks the source card images (409x600px) down to a size
	// that fits several tableau piles on screen.
	cardScale  = 0.25
	cardWidth  = 102 // 409 * cardScale, rounded
	cardHeight = 150 // 600 * cardScale, rounded

	stockOriginX = 16
	stockOriginY = 16
	wasteOriginX = stockOriginX + cardWidth + 16
	wasteOriginY = stockOriginY

	tableauOriginX = 16
	tableauOriginY = stockOriginY + cardHeight + 24
	tableauGapX    = cardWidth + 16
	faceUpOffsetY  = 24
)

// SolitaireScene draws the stock, waste, and tableau piles of a freshly
// dealt board, and lets the player click the stock to draw.
type SolitaireScene struct {
	board  *solitaire.Board
	images map[string]*ebiten.Image
	back   *ebiten.Image
}

// NewSolitaireScene deals a new, shuffled board and preloads the card
// images needed to draw it.
func NewSolitaireScene() (*SolitaireScene, error) {
	deck := card.NewDeck()
	card.Shuffle(deck, rand.New(rand.NewSource(time.Now().UnixNano())))

	board, err := solitaire.Deal(deck)
	if err != nil {
		return nil, err
	}

	var s SolitaireScene
	s.board = board
	s.images = make(map[string]*ebiten.Image)

	back, err := loadImage("cards/back.png")
	if err != nil {
		return nil, err
	}
	s.back = back

	for _, pile := range board.Tableau {
		for _, c := range pile.Cards {
			if _, err := s.cardImage(c); err != nil {
				return nil, err
			}
		}
	}
	for _, c := range board.Stock {
		if _, err := s.cardImage(c); err != nil {
			return nil, err
		}
	}

	return &s, nil
}

// cardImage returns the (cached) image for c, loading it on first use.
func (s *SolitaireScene) cardImage(c card.Card) (*ebiten.Image, error) {
	path := c.AssetPath()
	if img, ok := s.images[path]; ok {
		return img, nil
	}

	img, err := loadImage(path)
	if err != nil {
		return nil, err
	}

	s.images[path] = img
	return img, nil
}

func loadImage(path string) (*ebiten.Image, error) {
	f, err := assets.Assets.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(src), nil
}

// stockBounds returns the clickable rectangle of the stock pile.
func stockBounds() image.Rectangle {
	return image.Rect(stockOriginX, stockOriginY, stockOriginX+cardWidth, stockOriginY+cardHeight)
}

func (s *SolitaireScene) Update() (scene.Scene, error) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if image.Pt(x, y).In(stockBounds()) {
			s.board.DrawFromStock()
		}
	}

	return nil, nil
}

func (s *SolitaireScene) drawCard(screen *ebiten.Image, img *ebiten.Image, x, y int) {
	if img == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(cardScale, cardScale)
	op.GeoM.Translate(float64(x), float64(y))
	// The source images are much larger than their drawn size, so linear
	// filtering avoids the blocky look of the default nearest-neighbor
	// downscaling.
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(img, op)
}

func (s *SolitaireScene) Draw(screen *ebiten.Image) {
	if len(s.board.Stock) > 0 {
		s.drawCard(screen, s.back, stockOriginX, stockOriginY)
	}

	if top, ok := s.board.WasteTop(); ok {
		img, _ := s.cardImage(top)
		s.drawCard(screen, img, wasteOriginX, wasteOriginY)
	}

	for pileIndex, pile := range s.board.Tableau {
		x := tableauOriginX + pileIndex*tableauGapX
		faceDownCount := len(pile.Cards) - pile.FaceUp

		for i, c := range pile.Cards {
			y := tableauOriginY + i*faceUpOffsetY

			img := s.back
			if i >= faceDownCount {
				img, _ = s.cardImage(c)
			}
			s.drawCard(screen, img, x, y)
		}
	}
}
