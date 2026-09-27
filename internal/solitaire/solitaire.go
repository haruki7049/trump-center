// Package solitaire implements the board layout and move rules for
// Klondike solitaire, independent of any UI.
package solitaire

import (
	"fmt"

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
