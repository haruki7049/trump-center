// Package card provides the basic playing-card data model (suits, ranks,
// a standard 52-card deck) shared by any trump game built on top of it.
package card

import (
	"fmt"
	"math/rand"
)

// Suit is one of the four standard playing-card suits.
type Suit int

const (
	Spade Suit = iota
	Heart
	Diamond
	Club
)

// String returns the lowercase suit name, matching the directory names
// under assets/cards/ (e.g. "spade").
func (s Suit) String() string {
	switch s {
	case Spade:
		return "spade"
	case Heart:
		return "heart"
	case Diamond:
		return "diamond"
	case Club:
		return "club"
	default:
		return "unknown"
	}
}

// Red reports whether the suit is drawn in red (hearts and diamonds).
func (s Suit) Red() bool {
	return s == Heart || s == Diamond
}

// Rank is a card's rank, from Ace (1) to King (13).
type Rank int

const (
	Ace Rank = 1 + iota
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
)

// Card is a single playing card, identified by its suit and rank. Jokers
// are represented separately by Joker, since they have no suit or rank.
type Card struct {
	Suit Suit
	Rank Rank
}

// AssetPath returns the path of this card's image file, relative to the
// embedded assets root (see assets.Assets), e.g. "cards/spade/01.png".
func (c Card) AssetPath() string {
	return fmt.Sprintf("cards/%s/%02d.png", c.Suit, c.Rank)
}

// Joker is a special card outside the four suits.
type Joker struct{}

// AssetPath returns the path of the joker's image file, relative to the
// embedded assets root.
func (Joker) AssetPath() string {
	return "cards/joker.png"
}

// NewDeck returns the 52 standard playing cards (no jokers), ordered by
// suit then rank.
func NewDeck() []Card {
	suits := []Suit{Spade, Heart, Diamond, Club}
	deck := make([]Card, 0, len(suits)*int(King))

	for _, suit := range suits {
		for rank := Ace; rank <= King; rank++ {
			deck = append(deck, Card{Suit: suit, Rank: rank})
		}
	}

	return deck
}

// Shuffle randomises the order of deck in place using r.
func Shuffle(deck []Card, r *rand.Rand) {
	r.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}
