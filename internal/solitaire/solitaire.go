// Package solitaire implements the board layout and move rules for
// Klondike solitaire, independent of any UI.
package solitaire

import (
	"fmt"
	"slices"

	"github.com/haruki7049/trump-center/internal/card"
)

// TableauPileCount is the number of tableau piles in Klondike solitaire.
const TableauPileCount = 7

// Pile is an ordered stack of cards where the last element is the top of
// the pile. FaceUp counts how many cards, starting from the top, are
// face-up; the rest are face-down.
type Pile struct {
	Cards  []card.Card
	FaceUp int
}

// Top returns the card at the top of the pile, if any.
func (p Pile) Top() (card.Card, bool) {
	if len(p.Cards) == 0 {
		return card.Card{}, false
	}
	return p.Cards[len(p.Cards)-1], true
}

// Board is the state of a Klondike solitaire game.
type Board struct {
	Stock      []card.Card
	Waste      []card.Card
	Foundation [4][]card.Card
	Tableau    [TableauPileCount]Pile
}

// Deal builds a fresh Board from a standard, already-shuffled 52-card
// deck: tableau pile i (0-indexed) receives i+1 cards with only the top
// one face-up, and the remaining cards form the stock.
func Deal(deck []card.Card) (*Board, error) {
	if len(deck) != 52 {
		return nil, fmt.Errorf("solitaire: Deal requires 52 cards, got %d", len(deck))
	}

	var b Board
	i := 0
	for pileIndex := range TableauPileCount {
		size := pileIndex + 1
		b.Tableau[pileIndex] = Pile{
			Cards:  append([]card.Card(nil), deck[i:i+size]...),
			FaceUp: 1,
		}
		i += size
	}

	b.Stock = append([]card.Card(nil), deck[i:]...)
	return &b, nil
}

// DrawFromStock turns the top card of the stock face-up onto the waste
// pile. If the stock is empty, it instead recycles the waste back into
// the stock (in the order needed to reproduce the original draw order)
// without drawing a card, matching standard Klondike rules; the caller
// must call it again to draw.
func (b *Board) DrawFromStock() {
	if len(b.Stock) == 0 {
		for _, c := range slices.Backward(b.Waste) {
			b.Stock = append(b.Stock, c)
		}
		b.Waste = nil
		return
	}

	top := b.Stock[len(b.Stock)-1]
	b.Stock = b.Stock[:len(b.Stock)-1]
	b.Waste = append(b.Waste, top)
}

// WasteTop returns the card at the top of the waste pile, if any.
func (b *Board) WasteTop() (card.Card, bool) {
	if len(b.Waste) == 0 {
		return card.Card{}, false
	}
	return b.Waste[len(b.Waste)-1], true
}

// CanMoveToFoundation reports whether c can legally be placed on the
// foundation pile for its suit: an empty foundation only accepts an Ace,
// otherwise the next card must be one rank above the current top and of
// the same suit.
func (b *Board) CanMoveToFoundation(c card.Card) bool {
	pile := b.Foundation[c.Suit]
	if len(pile) == 0 {
		return c.Rank == card.Ace
	}

	top := pile[len(pile)-1]
	return c.Rank == top.Rank+1
}

// CanMoveToTableau reports whether c can legally be placed on top of the
// tableau pile at pileIndex: an empty pile only accepts a King, otherwise
// the next card must be one rank below the current top and of the
// opposite color.
func (b *Board) CanMoveToTableau(c card.Card, pileIndex int) bool {
	pile := b.Tableau[pileIndex]

	top, ok := pile.Top()
	if !ok {
		return c.Rank == card.King
	}

	return c.Rank == top.Rank-1 && c.Suit.Red() != top.Suit.Red()
}

// flipTopIfNeeded turns the new top card of a tableau pile face-up after
// its previous top card was removed, if it isn't already.
func (b *Board) flipTopIfNeeded(pileIndex int) {
	pile := &b.Tableau[pileIndex]
	if pile.FaceUp == 0 && len(pile.Cards) > 0 {
		pile.FaceUp = 1
	}
}

// MoveWasteToFoundation moves the top waste card onto its foundation
// pile, if legal, and reports whether the move happened.
func (b *Board) MoveWasteToFoundation() bool {
	c, ok := b.WasteTop()
	if !ok || !b.CanMoveToFoundation(c) {
		return false
	}

	b.Waste = b.Waste[:len(b.Waste)-1]
	b.Foundation[c.Suit] = append(b.Foundation[c.Suit], c)
	return true
}

// MoveWasteToTableau moves the top waste card onto the tableau pile at
// pileIndex, if legal, and reports whether the move happened.
func (b *Board) MoveWasteToTableau(pileIndex int) bool {
	c, ok := b.WasteTop()
	if !ok || !b.CanMoveToTableau(c, pileIndex) {
		return false
	}

	b.Waste = b.Waste[:len(b.Waste)-1]
	b.Tableau[pileIndex].Cards = append(b.Tableau[pileIndex].Cards, c)
	b.Tableau[pileIndex].FaceUp++
	return true
}

// MoveTableauToFoundation moves the face-up top card of the tableau pile
// at pileIndex onto its foundation pile, if legal, and reports whether
// the move happened.
func (b *Board) MoveTableauToFoundation(pileIndex int) bool {
	pile := &b.Tableau[pileIndex]

	c, ok := pile.Top()
	if !ok || pile.FaceUp == 0 || !b.CanMoveToFoundation(c) {
		return false
	}

	pile.Cards = pile.Cards[:len(pile.Cards)-1]
	pile.FaceUp--
	b.flipTopIfNeeded(pileIndex)
	b.Foundation[c.Suit] = append(b.Foundation[c.Suit], c)
	return true
}

// MoveTableauToTableau moves the face-up top card of the tableau pile at
// from onto the tableau pile at to, if legal, and reports whether the
// move happened.
func (b *Board) MoveTableauToTableau(from, to int) bool {
	if from == to {
		return false
	}

	fromPile := &b.Tableau[from]

	c, ok := fromPile.Top()
	if !ok || fromPile.FaceUp == 0 || !b.CanMoveToTableau(c, to) {
		return false
	}

	fromPile.Cards = fromPile.Cards[:len(fromPile.Cards)-1]
	fromPile.FaceUp--
	b.flipTopIfNeeded(from)

	b.Tableau[to].Cards = append(b.Tableau[to].Cards, c)
	b.Tableau[to].FaceUp++
	return true
}
