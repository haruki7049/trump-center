package solitaire

import (
	"image"

	"github.com/haruki7049/trump-center/internal/card"
	"github.com/haruki7049/trump-center/internal/solitaire"
)

// MediatorState is a state in the Mediator's drag/drop state machine.
type MediatorState int

const (
	// MediatorIdle is the state before a drag has started, and after one
	// ends.
	MediatorIdle MediatorState = iota
	// MediatorDragging is the state while the player is holding a card.
	MediatorDragging
)

// dragSource identifies where a card being dragged came from.
type dragSource int

const (
	dragSourceNone dragSource = iota
	dragSourceWaste
	dragSourceTableau
)

// dragInfo records what's being dragged, from where, and the offset from
// the cursor to its origin at the moment the drag started. cards holds
// every card being dragged as a unit, bottom of the run first; for a
// waste drag this is always a single card, but a tableau drag may carry a
// whole face-up run (see cardIndex, the run's position in its source
// pile, needed to apply the move on drop).
type dragInfo struct {
	source       dragSource
	tableauIndex int
	cardIndex    int
	cards        []card.Card
	offsetX      int
	offsetY      int
}

// DragState is a snapshot of an in-progress drag, returned by
// Mediator.Dragging so SolitaireScene can render the floating card(s) and
// skip drawing them at their source pile, without needing to know any of
// the Mediator's decision-making.
type DragState struct {
	Source       dragSource
	TableauIndex int
	CardIndex    int
	Cards        []card.Card
	OffsetX      int
	OffsetY      int
}

// Mediator is the terminal handler installed on the board's RootComponent
// (see Dispatch in event.go): every event that bubbles all the way up
// without being handled along the way reaches it. As a small state
// machine (Idle/Dragging), it owns all decision-making for the board —
// starting and ending a drag, and applying moves via
// internal/solitaire.Board — so that neither the Passive View components
// nor SolitaireScene need to know any game rules.
type Mediator struct {
	board *solitaire.Board
	root  *RootComponent

	stockPile      *PileComponent
	wastePile      *PileComponent
	foundationPile *PileComponent
	tableauPiles   [solitaire.TableauPileCount]*PileComponent

	state MediatorState
	drag  dragInfo

	rejected      RejectedDrop
	rejectedTicks int
}

// DropTargetKind identifies which kind of pile a RejectedDrop refers to.
type DropTargetKind int

const (
	// DropTargetFoundation is a foundation slot; RejectedDrop.Index is
	// its suit.
	DropTargetFoundation DropTargetKind = iota
	// DropTargetTableau is a tableau pile; RejectedDrop.Index is its
	// position.
	DropTargetTableau
)

// RejectedDrop identifies the pile a drag was just dropped on without the
// move being legal, so SolitaireScene can briefly mark it.
type RejectedDrop struct {
	Target DropTargetKind
	Index  int
}

// rejectedDropTicks is how many Update ticks (half a second at
// Ebitengine's default 60 TPS) a RejectedDrop stays reported after it
// happens.
const rejectedDropTicks = 30

// NewMediator builds a Mediator for board and the given piles, and
// installs it as root's event handler.
func NewMediator(
	board *solitaire.Board,
	root *RootComponent,
	stockPile, wastePile, foundationPile *PileComponent,
	tableauPiles [solitaire.TableauPileCount]*PileComponent,
) *Mediator {
	m := &Mediator{
		board:          board,
		root:           root,
		stockPile:      stockPile,
		wastePile:      wastePile,
		foundationPile: foundationPile,
		tableauPiles:   tableauPiles,
	}
	root.SetEventHandler(m.Handle)
	return m
}

// Handle is installed as Root's event handler (see RootComponent.HandleEvent);
// it is the state machine's single entry point.
func (m *Mediator) Handle(e Event) {
	switch e.Type {
	case EventPointerDown:
		m.handlePointerDown(e.X, e.Y)
	case EventPointerUp:
		m.handlePointerUp(e.X, e.Y)
	}
}

// Dragging reports the current DragState and whether a drag is active.
func (m *Mediator) Dragging() (DragState, bool) {
	if m.state != MediatorDragging {
		return DragState{}, false
	}
	return DragState{
		Source:       m.drag.source,
		TableauIndex: m.drag.tableauIndex,
		CardIndex:    m.drag.cardIndex,
		Cards:        m.drag.cards,
		OffsetX:      m.drag.offsetX,
		OffsetY:      m.drag.offsetY,
	}, true
}

// RejectedDrop reports the most recent illegal drop, and whether it is
// still recent enough (see rejectedDropTicks) to be shown.
func (m *Mediator) RejectedDrop() (RejectedDrop, bool) {
	if m.rejectedTicks <= 0 {
		return RejectedDrop{}, false
	}
	return m.rejected, true
}

// Tick advances time by one Update tick, letting a RejectedDrop expire.
func (m *Mediator) Tick() {
	if m.rejectedTicks > 0 {
		m.rejectedTicks--
	}
}

func (m *Mediator) reject(r RejectedDrop) {
	m.rejected = r
	m.rejectedTicks = rejectedDropTicks
}

// DropTargets reports, for an in-progress drag, which piles the dragged
// card(s) could legally be dropped on right now. Foundation is indexed by
// suit, the same way internal/solitaire.Board.Foundation is.
type DropTargets struct {
	Foundation [4]bool
	Tableau    [solitaire.TableauPileCount]bool
}

// ValidDropTargets reports where the current drag could legally be
// dropped, and whether a drag is active at all. It applies exactly the
// same rules as handlePointerUp, so SolitaireScene can highlight those
// piles without knowing any game rules itself.
func (m *Mediator) ValidDropTargets() (DropTargets, bool) {
	if m.state != MediatorDragging {
		return DropTargets{}, false
	}

	var t DropTargets
	bottom := m.drag.cards[0]

	switch m.drag.source {
	case dragSourceWaste:
		t.Foundation[bottom.Suit] = m.board.CanMoveToFoundation(bottom)
		for i := range t.Tableau {
			t.Tableau[i] = m.board.CanMoveToTableau(bottom, i)
		}
	case dragSourceTableau:
		// Only a single card can go to the foundation (see handlePointerUp).
		if len(m.drag.cards) == 1 {
			t.Foundation[bottom.Suit] = m.board.CanMoveToFoundation(bottom)
		}
		for i := range t.Tableau {
			t.Tableau[i] = m.board.CanMoveTableauRunToTableau(m.drag.tableauIndex, m.drag.cardIndex, i)
		}
	}

	return t, true
}

// handlePointerDown draws from the stock, or picks up the top card of the
// waste, or a face-up tableau card and every card above it as a run, if
// the pointer is over one of them. It does nothing while already
// dragging, or if the pointer isn't over anything pickable.
func (m *Mediator) handlePointerDown(x, y int) {
	if m.state == MediatorDragging {
		return
	}

	// Any new click means the player has moved on from the last
	// rejection.
	m.rejectedTicks = 0

	pt := image.Pt(x, y)

	if pt.In(m.stockPile.Bounds()) {
		m.board.DrawFromStock()
		return
	}

	if c, ok := m.board.WasteTop(); ok {
		if b, ok := m.wastePile.TopBounds(); ok && pt.In(b) {
			m.startDrag(dragSourceWaste, -1, -1, []card.Card{c}, x, y, b.Min.X, b.Min.Y)
			return
		}
	}

	for i, pile := range m.board.Tableau {
		faceDownCount := len(pile.Cards) - pile.FaceUp

		// Search from the top of the stack down to the first face-up
		// card: each card's bounds are full card-sized, so the
		// highest-index match is the one actually visible at (x, y).
		for cardIndex := len(pile.Cards) - 1; cardIndex >= faceDownCount; cardIndex-- {
			b, ok := m.tableauPiles[i].CardBoundsAt(cardIndex)
			if !ok || !pt.In(b) {
				continue
			}

			run := append([]card.Card(nil), pile.Cards[cardIndex:]...)
			m.startDrag(dragSourceTableau, i, cardIndex, run, x, y, b.Min.X, b.Min.Y)
			return
		}
	}
}

func (m *Mediator) startDrag(source dragSource, tableauIndex, cardIndex int, cards []card.Card, cursorX, cursorY, originX, originY int) {
	m.state = MediatorDragging
	m.drag = dragInfo{
		source:       source,
		tableauIndex: tableauIndex,
		cardIndex:    cardIndex,
		cards:        cards,
		offsetX:      cursorX - originX,
		offsetY:      cursorY - originY,
	}
}

// handlePointerUp attempts to move the dragged card onto whatever
// component is under (x, y). If that's a foundation or tableau pile and
// the move is illegal, it records a RejectedDrop; dropping anywhere else,
// including back onto the card's own tableau pile, just cancels the drag.
// It does nothing if no drag is in progress, and always ends the drag
// (returning to MediatorIdle) if one was.
func (m *Mediator) handlePointerUp(x, y int) {
	if m.state != MediatorDragging {
		return
	}
	defer func() {
		m.state = MediatorIdle
		m.drag = dragInfo{}
	}()

	switch hit := m.root.HitTest(x, y); hit {
	case Component(m.foundationPile):
		var ok bool
		switch m.drag.source {
		case dragSourceWaste:
			ok = m.board.MoveWasteToFoundation()
		case dragSourceTableau:
			// Only a single card can go to the foundation; a multi-card
			// run is always rejected there, matching standard Klondike
			// rules.
			if len(m.drag.cards) == 1 {
				ok = m.board.MoveTableauToFoundation(m.drag.tableauIndex)
			}
		}
		if !ok {
			m.reject(RejectedDrop{Target: DropTargetFoundation, Index: int(m.drag.cards[0].Suit)})
		}
	default:
		for i, p := range m.tableauPiles {
			if hit != Component(p) {
				continue
			}

			var ok bool
			switch m.drag.source {
			case dragSourceWaste:
				ok = m.board.MoveWasteToTableau(i)
			case dragSourceTableau:
				if i == m.drag.tableauIndex {
					return
				}
				ok = m.board.MoveTableauToTableauRun(m.drag.tableauIndex, m.drag.cardIndex, i)
			}
			if !ok {
				m.reject(RejectedDrop{Target: DropTargetTableau, Index: i})
			}
			return
		}
	}
}
