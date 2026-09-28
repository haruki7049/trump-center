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

	"github.com/ebitenui/ebitenui"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/haruki7049/trump-center/assets"
	"github.com/haruki7049/trump-center/internal/card"
	"github.com/haruki7049/trump-center/internal/scene"
	"github.com/haruki7049/trump-center/internal/solitaire"
)

const (
	// cardScale shrinks the source card images (409x600px) down to a size
	// that fits several tableau piles on screen. 0.3 is about the largest
	// that still keeps the top row (stock, waste, four foundations) clear
	// of the "New Game" button in a 1280px-wide window.
	cardScale  = 0.3
	cardWidth  = 123 // 409 * cardScale, rounded
	cardHeight = 180 // 600 * cardScale, rounded

	// stockOriginY leaves labelOffsetY of headroom above the first row for
	// its area labels (see drawLabels).
	stockOriginX = 16
	stockOriginY = 60
	// stockMaxLayers/stockLayerOffsetX/Y draw a few back-of-card images
	// peeking out from behind the front card to suggest a stack with some
	// thickness, instead of a single flat card. The front (clickable) card
	// is always drawn last, at (stockOriginX, stockOriginY) exactly, so
	// stockPile's hit-test bounds don't need to change to accommodate this.
	stockMaxLayers   = 4
	stockLayerOffset = 3
	wasteOriginX     = stockOriginX + cardWidth + 16
	wasteOriginY     = stockOriginY
	// wasteFanCount/wasteFanOffsetX fan the most recent waste cards out
	// horizontally instead of showing only a single flat card, so the
	// player can see a little of their recent draw history. Only the
	// rightmost (most recent) card is ever pickable.
	wasteFanCount   = 3
	wasteFanOffsetX = 20

	foundationOriginX = wasteOriginX + cardWidth + (wasteFanCount-1)*wasteFanOffsetX + 24
	foundationOriginY = stockOriginY
	foundationCount   = 4
	foundationGapX    = cardWidth + 16

	tableauOriginX = 16
	// tableauOriginY leaves a bit more room than the bare minimum below
	// the first row, so the "Tableau" label also fits above it.
	tableauOriginY = stockOriginY + cardHeight + 60
	tableauGapX    = cardWidth + 16
	// faceDownOffsetY and faceUpOffsetY are the vertical overlap steps in
	// a tableau pile. Face-down cards show nothing worth reading, so they
	// overlap more tightly, leaving room for the face-up cards' corner
	// indices. faceUpOffsetY also spaces a dragged run.
	faceDownOffsetY = 12
	faceUpOffsetY   = 28
	// minFaceUpOffsetY is how tightly a tall pile's face-up cards may be
	// squeezed to keep it above tableauMaxBottom (see tableauCardY).
	minFaceUpOffsetY = 12
	// tableauMaxBottom is the lowest Y a tableau pile's last card should
	// reach: the window height (internal/game.WINDOW_HEIGHT, hardcoded for
	// the same import-cycle reason as winMessageX/Y) minus a small margin.
	tableauMaxBottom = 720 - 16

	// tableauColumnHeight is an arbitrarily generous height used only to
	// detect drops anywhere below a tableau pile's origin.
	tableauColumnHeight = 2000

	// labelFontSize and labelOffsetY size and position the small area
	// labels drawn above the stock, waste, foundation, and tableau.
	labelFontSize = 28
	labelOffsetY  = 40

	stockLabel      = "Stock"
	wasteLabel      = "Waste"
	foundationLabel = "Foundation"
	tableauLabel    = "Tableau"

	// winMessageText is shown once Board.Won reports the game complete.
	// winMessageX/Y are hardcoded rather than centered on the window size
	// (internal/game.WINDOW_WIDTH/HEIGHT) to avoid an import cycle
	// (internal/game -> title -> solitaire).
	winMessageText = "You win!"
	winMessageX    = 640
	winMessageY    = 300

	// stuckMessageText/stuckHintText are shown once Mediator.Stuck reports
	// the game can no longer be won, over a translucent band
	// (stuckBandY/stuckBandHeight, spanning windowWidth) so they stay
	// readable on top of the tableau. Like winMessageX/Y, these are
	// hardcoded rather than read from internal/game to avoid an import
	// cycle.
	stuckMessageText = "No moves left"
	stuckHintText    = "Press New Game to start over"
	windowWidth      = 1280
	stuckBandY       = 320
	stuckBandHeight  = 130
	stuckMessageY    = stuckBandY + 12
	stuckHintY       = stuckMessageY + 64
)

var (
	// validDropColor outlines piles the current drag could be dropped on.
	validDropColor = color.NRGBA{0xff, 0xd7, 0x00, 0xff}
	// rejectedDropColor briefly outlines a pile a drag was just illegally
	// dropped on (see Mediator.RejectedDrop).
	rejectedDropColor = color.NRGBA{0xff, 0x30, 0x30, 0xff}
	// stuckBandColor is the translucent band behind the "No moves left"
	// message.
	stuckBandColor = color.NRGBA{0x00, 0x00, 0x00, 0xd0}
)

// SolitaireScene draws the stock, waste, foundation, and tableau piles of
// a freshly dealt board, and lets the player draw from the stock and
// drag cards between piles. All decision-making for those interactions
// lives in its Mediator; SolitaireScene itself only polls input,
// dispatches events, and draws.
type SolitaireScene struct {
	board         *solitaire.Board
	images        map[string]*ebiten.Image
	back          *ebiten.Image
	fontFace      text.Face
	labelFontFace text.Face
	ui            *ebitenui.UI

	root           *RootComponent
	stockPile      *PileComponent
	wastePile      *PileComponent
	foundationPile *PileComponent
	tableauPiles   [solitaire.TableauPileCount]*PileComponent
	mediator       *Mediator

	newGameRequested bool
}

// NewSolitaireScene deals a new, shuffled board and preloads the assets
// needed to draw it and its "New Game" button.
func NewSolitaireScene() (*SolitaireScene, error) {
	var s SolitaireScene
	s.images = make(map[string]*ebiten.Image)

	back, err := loadImage("cards/back.png")
	if err != nil {
		return nil, err
	}
	s.back = back

	fontFace, err := loadFont("fonts/DotGothic16/DotGothic16-Regular.ttf", 48)
	if err != nil {
		return nil, err
	}
	s.fontFace = fontFace

	labelFontFace, err := loadFont("fonts/DotGothic16/DotGothic16-Regular.ttf", labelFontSize)
	if err != nil {
		return nil, err
	}
	s.labelFontFace = labelFontFace

	if err := s.deal(); err != nil {
		return nil, err
	}

	s.buildUI()

	return &s, nil
}

// deal shuffles a fresh 52-card deck, deals it into a new Board, and
// rebuilds the component tree and Mediator around it. It is used both by
// NewSolitaireScene at startup and by Update whenever the player clicks
// "New Game".
func (s *SolitaireScene) deal() error {
	deck := card.NewDeck()
	card.Shuffle(deck, rand.New(rand.NewSource(time.Now().UnixNano())))

	board, err := solitaire.Deal(deck)
	if err != nil {
		return err
	}
	s.board = board

	for _, pile := range board.Tableau {
		for _, c := range pile.Cards {
			if _, err := s.cardImage(c); err != nil {
				return err
			}
		}
	}
	for _, c := range board.Stock {
		if _, err := s.cardImage(c); err != nil {
			return err
		}
	}

	s.stockPile = NewPileComponent(image.Rect(
		stockOriginX, stockOriginY, stockOriginX+cardWidth, stockOriginY+cardHeight,
	))
	s.wastePile = NewPileComponent(image.Rect(
		wasteOriginX, wasteOriginY,
		wasteOriginX+cardWidth+(wasteFanCount-1)*wasteFanOffsetX, wasteOriginY+cardHeight,
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

	return nil
}

// buildUI builds the scene's narrow-sense UI: a single "New Game" button,
// anchored to the top-right corner so it never overlaps the board. This
// is real interactive chrome (unlike the plain win-message text drawn
// directly in Draw), so per the project's UI policy it goes through
// ebitenui, coexisting here with the custom Composite/Mediator board
// architecture on the same scene.
func (s *SolitaireScene) buildUI() {
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	newGameButton := widget.NewButton(
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    eimage.NewNineSliceColor(color.NRGBA{0x40, 0x40, 0x40, 0xff}),
			Hover:   eimage.NewNineSliceColor(color.NRGBA{0x60, 0x60, 0x60, 0xff}),
			Pressed: eimage.NewNineSliceColor(color.NRGBA{0x20, 0x20, 0x20, 0xff}),
		}),
		widget.ButtonOpts.Text("New Game", &s.fontFace, &widget.ButtonTextColor{Idle: color.White}),
		widget.ButtonOpts.TextPadding(&widget.Insets{Left: 16, Right: 16, Top: 8, Bottom: 8}),
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionEnd,
			VerticalPosition:   widget.AnchorLayoutPositionStart,
			Padding:            &widget.Insets{Top: 16, Right: 16},
		})),
		widget.ButtonOpts.ClickedHandler(func(_ *widget.ButtonClickedEventArgs) {
			s.newGameRequested = true
		}),
	)
	rootContainer.AddChild(newGameButton)

	s.ui = &ebitenui.UI{Container: rootContainer}
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

func loadFont(path string, size float64) (*text.GoTextFace, error) {
	f, err := assets.Assets.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	src, err := text.NewGoTextFaceSource(f)
	if err != nil {
		return nil, err
	}

	return &text.GoTextFace{Source: src, Size: size}, nil
}

// Update polls input and dispatches it as Events, which bubble up to the
// Mediator (installed on s.root) via Chain of Responsibility; the
// Mediator owns all decisions about what those events mean. It also
// drives the "New Game" button's ebitenui.UI, which lives alongside but
// outside that event chain, same as the title scene's own UI.
func (s *SolitaireScene) Update() (scene.Scene, error) {
	s.ui.Update()
	s.mediator.Tick()

	if s.newGameRequested {
		s.newGameRequested = false
		if err := s.deal(); err != nil {
			return nil, err
		}
		return nil, nil
	}

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

// drawCenteredText draws str in white, horizontally centered on the
// window (at winMessageX), with its top at y.
func drawCenteredText(screen *ebiten.Image, str string, face text.Face, y float64) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(winMessageX, y)
	op.PrimaryAlign = text.AlignCenter
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, str, face, op)
}

// drawLabels draws a small caption above each of the four play areas
// (stock, waste, foundation, tableau), so a player unfamiliar with
// Klondike's layout conventions can identify them at a glance. The
// tableau gets a single label above its first column, representing the
// whole row of seven piles as one area.
func (s *SolitaireScene) drawLabel(screen *ebiten.Image, label string, x, y int) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, label, s.labelFontFace, op)
}

func (s *SolitaireScene) drawLabels(screen *ebiten.Image) {
	s.drawLabel(screen, stockLabel, stockOriginX, stockOriginY-labelOffsetY)
	s.drawLabel(screen, wasteLabel, wasteOriginX, wasteOriginY-labelOffsetY)
	s.drawLabel(screen, foundationLabel, foundationOriginX, foundationOriginY-labelOffsetY)
	s.drawLabel(screen, tableauLabel, tableauOriginX, tableauOriginY-labelOffsetY)
}

// syncComponents recomputes the CardDraw values for every PileComponent
// from the current board state, so the (Passive View) component tree
// reflects the latest game state before it is drawn.
func (s *SolitaireScene) syncComponents() {
	// Draw a few back-of-card layers peeking out to the upper-left of the
	// front card to suggest a stack with some thickness. Drawn back to
	// front, so the front (clickable) layer is always last and lands
	// exactly at (stockOriginX, stockOriginY).
	var stockCards []CardDraw
	if layers := min(len(s.board.Stock), stockMaxLayers); layers > 0 {
		for i := range layers {
			depth := layers - 1 - i
			stockCards = append(stockCards, CardDraw{
				Image: s.back,
				X:     stockOriginX - depth*stockLayerOffset,
				Y:     stockOriginY - depth*stockLayerOffset,
			})
		}
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

	// Fan out the last wasteFanCount cards, oldest to the left, so the
	// player sees a little of their recent draw history; only the
	// rightmost (last) one is ever pickable (see Mediator). While it's
	// being dragged, drop it from the fan so the card(s) behind it show
	// through, the same way the tableau reveals cards behind a dragged run.
	waste := s.board.Waste
	draggingWaste := dragging && dragState.Source == dragSourceWaste
	if draggingWaste && len(waste) > 0 {
		waste = waste[:len(waste)-1]
	}
	shown := min(len(waste), wasteFanCount)
	waste = waste[len(waste)-shown:]

	var wasteCards []CardDraw
	for i, c := range waste {
		img, _ := s.cardImage(c)
		wasteCards = append(wasteCards, CardDraw{Image: img, X: wasteOriginX + i*wasteFanOffsetX, Y: wasteOriginY})
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
			cards = append(cards, CardDraw{Image: img, X: x, Y: tableauCardY(faceDownCount, pile.FaceUp, i)})
		}
		s.tableauPiles[pileIndex].SetCards(cards)
	}

	s.syncHighlights()
}

// tableauCardY returns the Y position of the card at index i in a tableau
// pile with faceDownCount face-down cards under faceUpCount face-up ones.
// Face-down cards step by faceDownOffsetY and face-up ones by
// faceUpOffsetY, except that the face-up step shrinks (down to
// minFaceUpOffsetY) when needed to keep the pile's last card above
// tableauMaxBottom. The step depends only on the pile's contents, not on
// what is currently being dragged out of it, so cards don't shift while
// a run is lifted.
func tableauCardY(faceDownCount, faceUpCount, i int) int {
	if i < faceDownCount {
		return tableauOriginY + i*faceDownOffsetY
	}

	faceUpTop := tableauOriginY + faceDownCount*faceDownOffsetY
	step := faceUpOffsetY
	if faceUpCount > 1 {
		room := tableauMaxBottom - cardHeight - faceUpTop
		step = max(minFaceUpOffsetY, min(faceUpOffsetY, room/(faceUpCount-1)))
	}
	return faceUpTop + (i-faceDownCount)*step
}

// syncHighlights outlines, as reported by the Mediator, every pile the
// current drag could legally be dropped on (validDropColor), and the pile
// a drag was just illegally dropped on (rejectedDropColor); with neither,
// all outlines are cleared. Each outline surrounds the card a drop would
// land on, or the empty slot if the pile is empty.
func (s *SolitaireScene) syncHighlights() {
	targets, _ := s.mediator.ValidDropTargets()
	rejected, hasRejected := s.mediator.RejectedDrop()

	var foundationHighlights []Highlight
	for i, ok := range targets.Foundation {
		isRejected := hasRejected && rejected.Target == DropTargetFoundation && rejected.Index == i
		if !ok && !isRejected {
			continue
		}
		x := foundationOriginX + i*foundationGapX
		foundationHighlights = append(foundationHighlights, Highlight{
			Rect:  image.Rect(x, foundationOriginY, x+cardWidth, foundationOriginY+cardHeight),
			Color: highlightColorFor(isRejected),
		})
	}
	s.foundationPile.SetHighlights(foundationHighlights)

	for i, ok := range targets.Tableau {
		pile := s.tableauPiles[i]
		isRejected := hasRejected && rejected.Target == DropTargetTableau && rejected.Index == i
		if !ok && !isRejected {
			pile.SetHighlights(nil)
			continue
		}

		r, hasTop := pile.TopBounds()
		if !hasTop {
			x := tableauOriginX + i*tableauGapX
			r = image.Rect(x, tableauOriginY, x+cardWidth, tableauOriginY+cardHeight)
		}
		pile.SetHighlights([]Highlight{{Rect: r, Color: highlightColorFor(isRejected)}})
	}
}

// highlightColorFor picks rejectedDropColor for a rejected drop's pile,
// and validDropColor otherwise.
func highlightColorFor(rejected bool) color.Color {
	if rejected {
		return rejectedDropColor
	}
	return validDropColor
}

func (s *SolitaireScene) Draw(screen *ebiten.Image) {
	s.syncComponents()
	s.drawLabels(screen)
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
		drawCenteredText(screen, winMessageText, s.fontFace, winMessageY)
	} else if s.mediator.Stuck() {
		vector.FillRect(screen, 0, stuckBandY, windowWidth, stuckBandHeight, stuckBandColor, false)
		drawCenteredText(screen, stuckMessageText, s.fontFace, stuckMessageY)
		drawCenteredText(screen, stuckHintText, s.labelFontFace, stuckHintY)
	}

	s.ui.Draw(screen)
}
