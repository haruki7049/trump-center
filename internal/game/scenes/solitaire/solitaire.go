// Package solitaire is the scene that draws a Klondike solitaire board.
// Cards can be drawn from the stock by clicking it, and the top card of
// the waste or of any tableau pile can be dragged onto a tableau pile or
// the foundation row.
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

	foundationOriginX = wasteOriginX + cardWidth + 48
	foundationOriginY = stockOriginY
	foundationCount   = 4
	foundationGapX    = cardWidth + 16

	tableauOriginX = 16
	tableauOriginY = stockOriginY + cardHeight + 24
	tableauGapX    = cardWidth + 16
	faceUpOffsetY  = 24

	// tableauColumnHeight is an arbitrarily generous height used only to
	// detect drops anywhere below a tableau pile's origin.
	tableauColumnHeight = 2000
)

// dragSource identifies where a card being dragged came from.
type dragSource int

const (
	dragSourceNone dragSource = iota
	dragSourceWaste
	dragSourceTableau
)

// drag tracks the card currently being dragged by the player, if any.
type drag struct {
	active       bool
	source       dragSource
	tableauIndex int
	card         card.Card
	offsetX      int
	offsetY      int
	cursorX      int
	cursorY      int
}

// SolitaireScene draws the stock, waste, foundation, and tableau piles of
// a freshly dealt board, and lets the player draw from the stock and
// drag cards between piles.
type SolitaireScene struct {
	board  *solitaire.Board
	images map[string]*ebiten.Image
	back   *ebiten.Image
	drag   drag

	root           *RootComponent
	stockPile      *PileComponent
	wastePile      *PileComponent
	foundationPile *PileComponent
	tableauPiles   [solitaire.TableauPileCount]*PileComponent
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

	s.stockPile = NewPileComponent(image.Rect(
		stockOriginX, stockOriginY, stockOriginX+cardWidth, stockOriginY+cardHeight,
	))
	s.wastePile = NewPileComponent(image.Rect(
		wasteOriginX, wasteOriginY, wasteOriginX+cardWidth, wasteOriginY+cardHeight,
	))
	s.foundationPile = NewPileComponent(image.Rect(
		foundationOriginX, foundationOriginY,
		foundationOriginX+foundationCount*foundationGapX-16, foundationOriginY+cardHeight,
	))
	for i := range s.tableauPiles {
		x := tableauOriginX + i*tableauGapX
		s.tableauPiles[i] = NewPileComponent(image.Rect(
			x, tableauOriginY, x+cardWidth, tableauOriginY+tableauColumnHeight,
		))
	}

	children := []Component{s.stockPile, s.wastePile, s.foundationPile}
	for _, p := range s.tableauPiles {
		children = append(children, p)
	}
	s.root = NewRootComponent(children...)

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

func (s *SolitaireScene) startDrag(source dragSource, tableauIndex int, c card.Card, cursorX, cursorY, originX, originY int) {
	s.drag = drag{
		active:       true,
		source:       source,
		tableauIndex: tableauIndex,
		card:         c,
		offsetX:      cursorX - originX,
		offsetY:      cursorY - originY,
		cursorX:      cursorX,
		cursorY:      cursorY,
	}
}

// dropDrag attempts to move the dragged card onto whatever component is
// under (x, y), doing nothing if the drop location or move is invalid.
func (s *SolitaireScene) dropDrag(x, y int) {
	switch hit := s.root.HitTest(x, y); hit {
	case Component(s.foundationPile):
		switch s.drag.source {
		case dragSourceWaste:
			s.board.MoveWasteToFoundation()
		case dragSourceTableau:
			s.board.MoveTableauToFoundation(s.drag.tableauIndex)
		}
	default:
		for i, p := range s.tableauPiles {
			if hit != Component(p) {
				continue
			}

			switch s.drag.source {
			case dragSourceWaste:
				s.board.MoveWasteToTableau(i)
			case dragSourceTableau:
				if i != s.drag.tableauIndex {
					s.board.MoveTableauToTableau(s.drag.tableauIndex, i)
				}
			}
			return
		}
	}
}

func (s *SolitaireScene) Update() (scene.Scene, error) {
	// Keep the (Passive View) component tree's TopBounds current before
	// hit-testing against it below.
	s.syncComponents()

	if !s.drag.active {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			s.tryStartDrag(x, y)
		}
		return nil, nil
	}

	x, y := ebiten.CursorPosition()
	s.drag.cursorX, s.drag.cursorY = x, y

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		s.dropDrag(x, y)
		s.drag = drag{}
	}

	return nil, nil
}

// tryStartDrag handles a fresh left-click at (x, y): drawing from the
// stock, or picking up the top card of the waste or a tableau pile.
func (s *SolitaireScene) tryStartDrag(x, y int) {
	pt := image.Pt(x, y)

	if pt.In(s.stockPile.Bounds()) {
		s.board.DrawFromStock()
		return
	}

	if c, ok := s.board.WasteTop(); ok {
		if b, ok := s.wastePile.TopBounds(); ok && pt.In(b) {
			s.startDrag(dragSourceWaste, -1, c, x, y, b.Min.X, b.Min.Y)
			return
		}
	}

	for i, pile := range s.board.Tableau {
		if pile.FaceUp == 0 {
			continue
		}

		c, ok := pile.Top()
		if !ok {
			continue
		}

		b, ok := s.tableauPiles[i].TopBounds()
		if ok && pt.In(b) {
			s.startDrag(dragSourceTableau, i, c, x, y, b.Min.X, b.Min.Y)
			return
		}
	}
}

// drawCard renders a single card image at (x, y) in board coordinates,
// scaled down and linearly filtered to avoid the blocky look of the
// default nearest-neighbor downscaling.
func drawCard(screen *ebiten.Image, img *ebiten.Image, x, y int) {
	if img == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(cardScale, cardScale)
	op.GeoM.Translate(float64(x), float64(y))
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(img, op)
}

// syncComponents recomputes the CardDraw values for every PileComponent
// from the current board state, so the (Passive View) component tree
// reflects the latest game state before it is drawn.
func (s *SolitaireScene) syncComponents() {
	var stockCards []CardDraw
	if len(s.board.Stock) > 0 {
		stockCards = []CardDraw{{Image: s.back, X: stockOriginX, Y: stockOriginY}}
	}
	s.stockPile.SetCards(stockCards)

	var foundationCards []CardDraw
	for i, pile := range s.board.Foundation {
		if len(pile) == 0 {
			continue
		}
		img, _ := s.cardImage(pile[len(pile)-1])
		foundationCards = append(foundationCards, CardDraw{
			Image: img,
			X:     foundationOriginX + i*foundationGapX,
			Y:     foundationOriginY,
		})
	}
	s.foundationPile.SetCards(foundationCards)

	var wasteCards []CardDraw
	draggingWaste := s.drag.active && s.drag.source == dragSourceWaste
	if top, ok := s.board.WasteTop(); ok && !draggingWaste {
		img, _ := s.cardImage(top)
		wasteCards = []CardDraw{{Image: img, X: wasteOriginX, Y: wasteOriginY}}
	}
	s.wastePile.SetCards(wasteCards)

	for pileIndex, pile := range s.board.Tableau {
		x := tableauOriginX + pileIndex*tableauGapX
		faceDownCount := len(pile.Cards) - pile.FaceUp
		draggingHere := s.drag.active && s.drag.source == dragSourceTableau && s.drag.tableauIndex == pileIndex

		var cards []CardDraw
		for i, c := range pile.Cards {
			if draggingHere && i == len(pile.Cards)-1 {
				continue
			}

			img := s.back
			if i >= faceDownCount {
				img, _ = s.cardImage(c)
			}
			cards = append(cards, CardDraw{Image: img, X: x, Y: tableauOriginY + i*faceUpOffsetY})
		}
		s.tableauPiles[pileIndex].SetCards(cards)
	}
}

func (s *SolitaireScene) Draw(screen *ebiten.Image) {
	s.syncComponents()
	s.root.Draw(screen)

	if s.drag.active {
		img, _ := s.cardImage(s.drag.card)
		drawCard(screen, img, s.drag.cursorX-s.drag.offsetX, s.drag.cursorY-s.drag.offsetY)
	}
}
