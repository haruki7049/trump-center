// Package solitaire is the scene that draws a Klondike solitaire board.
// Cards can be drawn from the stock by clicking it, and the top card of
// the waste or of any tableau pile can be dragged onto a tableau pile or
// the foundation row.
package solitaire

import (
	"image"
	"image/color"
	_ "image/png"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
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

	// winMessageText is shown once Board.Won reports the game complete.
	// winMessageX/Y are hardcoded rather than centered on the window size
	// (internal/game.WINDOW_WIDTH/HEIGHT) to avoid an import cycle
	// (internal/game -> title -> solitaire).
	winMessageText = "You win!"
	winMessageX    = 640
	winMessageY    = 300
)

// SolitaireScene draws the stock, waste, foundation, and tableau piles of
// a freshly dealt board, and lets the player draw from the stock and
// drag cards between piles. All decision-making for those interactions
// lives in its Mediator; SolitaireScene itself only polls input,
// dispatches events, and draws.
type SolitaireScene struct {
	board    *solitaire.Board
	images   map[string]*ebiten.Image
	back     *ebiten.Image
	fontFace text.Face

	root           *RootComponent
	stockPile      *PileComponent
	wastePile      *PileComponent
	foundationPile *PileComponent
	tableauPiles   [solitaire.TableauPileCount]*PileComponent
	mediator       *Mediator
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

	fontFace, err := loadFont("fonts/DotGothic16/DotGothic16-Regular.ttf")
	if err != nil {
		return nil, err
	}
	s.fontFace = fontFace

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

	s.mediator = NewMediator(s.board, s.root, s.stockPile, s.wastePile, s.foundationPile, s.tableauPiles)

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

func loadFont(path string) (*text.GoTextFace, error) {
	f, err := assets.Assets.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	src, err := text.NewGoTextFaceSource(f)
	if err != nil {
		return nil, err
	}

	return &text.GoTextFace{Source: src, Size: 48}, nil
}

// Update polls input and dispatches it as Events, which bubble up to the
// Mediator (installed on s.root) via Chain of Responsibility; the
// Mediator owns all decisions about what those events mean.
func (s *SolitaireScene) Update() (scene.Scene, error) {
	// Keep the (Passive View) component tree's TopBounds current before
	// the Mediator hit-tests against it while handling the events below.
	s.syncComponents()

	x, y := ebiten.CursorPosition()

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if hit := s.root.HitTest(x, y); hit != nil {
			Dispatch(hit, Event{Type: EventPointerDown, X: x, Y: y})
		}
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		Dispatch(s.root, Event{Type: EventPointerUp, X: x, Y: y})
	}

	return nil, nil
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

	dragState, dragging := s.mediator.Dragging()

	var wasteCards []CardDraw
	draggingWaste := dragging && dragState.Source == dragSourceWaste
	if top, ok := s.board.WasteTop(); ok && !draggingWaste {
		img, _ := s.cardImage(top)
		wasteCards = []CardDraw{{Image: img, X: wasteOriginX, Y: wasteOriginY}}
	}
	s.wastePile.SetCards(wasteCards)

	for pileIndex, pile := range s.board.Tableau {
		x := tableauOriginX + pileIndex*tableauGapX
		faceDownCount := len(pile.Cards) - pile.FaceUp
		draggingHere := dragging && dragState.Source == dragSourceTableau && dragState.TableauIndex == pileIndex

		var cards []CardDraw
		for i, c := range pile.Cards {
			if draggingHere && i >= dragState.CardIndex {
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

	if dragState, ok := s.mediator.Dragging(); ok {
		x, y := ebiten.CursorPosition()
		baseX, baseY := x-dragState.OffsetX, y-dragState.OffsetY
		for i, c := range dragState.Cards {
			img, _ := s.cardImage(c)
			drawCard(screen, img, baseX, baseY+i*faceUpOffsetY)
		}
	}

	if s.board.Won() {
		op := &text.DrawOptions{}
		op.GeoM.Translate(winMessageX, winMessageY)
		op.PrimaryAlign = text.AlignCenter
		op.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, winMessageText, s.fontFace, op)
	}
}
