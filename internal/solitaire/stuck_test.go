package solitaire_test

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/haruki7049/trump-center/internal/card"
	"github.com/haruki7049/trump-center/internal/solitaire"
)

func run(suit card.Suit, from, to card.Rank) []card.Card {
	var cards []card.Card
	for r := from; r <= to; r++ {
		cards = append(cards, card.Card{Suit: suit, Rank: r})
	}
	return cards
}

// reportedStuckBoard is the real position that motivated #79: stock and
// waste empty, 26 cards on the foundation, and the four face-down cards
// (8♠, 10♦, J♦, 5♣) each locked behind a card whose only possible
// destinations are those same face-down cards or cards beneath itself.
func reportedStuckBoard() *solitaire.Board {
	c := func(s card.Suit, r card.Rank) card.Card { return card.Card{Suit: s, Rank: r} }
	return &solitaire.Board{
		Foundation: [4][]card.Card{
			card.Spade:   run(card.Spade, card.Ace, card.Seven),
			card.Heart:   run(card.Heart, card.Ace, card.Nine),
			card.Diamond: run(card.Diamond, card.Ace, card.Six),
			card.Club:    run(card.Club, card.Ace, card.Four),
		},
		Tableau: [solitaire.TableauPileCount]solitaire.Pile{
			{Cards: []card.Card{
				c(card.Heart, card.King), c(card.Club, card.Queen), c(card.Heart, card.Jack), c(card.Spade, card.Ten),
				c(card.Diamond, card.Nine), c(card.Club, card.Eight), c(card.Diamond, card.Seven), c(card.Club, card.Six),
			}, FaceUp: 8},
			{Cards: []card.Card{c(card.Spade, card.King), c(card.Diamond, card.Queen), c(card.Club, card.Jack)}, FaceUp: 3},
			{Cards: []card.Card{
				c(card.Club, card.King), c(card.Heart, card.Queen), c(card.Spade, card.Jack), c(card.Heart, card.Ten),
				c(card.Club, card.Nine), c(card.Diamond, card.Eight), c(card.Club, card.Seven),
			}, FaceUp: 7},
			{Cards: []card.Card{c(card.Diamond, card.King), c(card.Spade, card.Queen)}, FaceUp: 2},
			{Cards: []card.Card{
				c(card.Club, card.Five), c(card.Spade, card.Eight), c(card.Diamond, card.Ten), // face-down
				c(card.Club, card.Ten),
			}, FaceUp: 1},
			{},
			{Cards: []card.Card{c(card.Diamond, card.Jack) /* face-down */, c(card.Spade, card.Nine)}, FaceUp: 1},
		},
	}
}

func TestBoard_Stuck_ReportedPosition(t *testing.T) {
	b := reportedStuckBoard()

	// Guard the test data itself: all 52 cards, each exactly once.
	seen := map[card.Card]bool{}
	for _, pile := range b.Foundation {
		for _, c := range pile {
			seen[c] = true
		}
	}
	for _, pile := range b.Tableau {
		for _, c := range pile.Cards {
			seen[c] = true
		}
	}
	if len(seen) != 52 {
		t.Fatalf("test board has %d distinct cards; want 52", len(seen))
	}

	if !b.Stuck() {
		t.Error("expected the reported position to be stuck")
	}
}

func TestBoard_Stuck(t *testing.T) {
	c := func(s card.Suit, r card.Rank) card.Card { return card.Card{Suit: s, Rank: r} }

	// Tableau piles 1-6 each hold a lone King: nothing can go on or move
	// them, so pile 0 and the stock/waste decide the outcome.
	withKings := func(b *solitaire.Board) *solitaire.Board {
		for i := 1; i < solitaire.TableauPileCount; i++ {
			b.Tableau[i] = solitaire.Pile{Cards: []card.Card{c(card.Club, card.King)}, FaceUp: 1}
		}
		return b
	}

	tests := []struct {
		name  string
		board *solitaire.Board
		want  bool
	}{
		{
			name: "unplayable stock card and a blocked face-down card",
			board: withKings(&solitaire.Board{
				Stock: []card.Card{c(card.Diamond, card.Five)},
				Tableau: [solitaire.TableauPileCount]solitaire.Pile{
					{Cards: []card.Card{c(card.Spade, card.Ace), c(card.Heart, card.Two)}, FaceUp: 1},
				},
			}),
			want: true,
		},
		{
			name: "playable card deep in the stock",
			board: withKings(&solitaire.Board{
				Stock: []card.Card{c(card.Diamond, card.Ace), c(card.Diamond, card.Five), c(card.Diamond, card.Six)},
				Tableau: [solitaire.TableauPileCount]solitaire.Pile{
					{Cards: []card.Card{c(card.Spade, card.Ace), c(card.Heart, card.Two)}, FaceUp: 1},
				},
			}),
			want: false,
		},
		{
			name: "playable card only after recycling the waste",
			board: withKings(&solitaire.Board{
				Waste: []card.Card{c(card.Diamond, card.Ace), c(card.Diamond, card.Five)},
				Tableau: [solitaire.TableauPileCount]solitaire.Pile{
					{Cards: []card.Card{c(card.Spade, card.Ace), c(card.Heart, card.Two)}, FaceUp: 1},
				},
			}),
			want: false,
		},
		{
			name: "face-down card freed by a tableau move",
			board: &solitaire.Board{
				Tableau: [solitaire.TableauPileCount]solitaire.Pile{
					{Cards: []card.Card{c(card.Spade, card.Two), c(card.Heart, card.Seven)}, FaceUp: 1},
					{Cards: []card.Card{c(card.Spade, card.Eight)}, FaceUp: 1},
				},
			},
			want: false,
		},
		{
			name: "won game",
			board: &solitaire.Board{Foundation: [4][]card.Card{
				run(card.Spade, card.Ace, card.King), run(card.Heart, card.Ace, card.King),
				run(card.Diamond, card.Ace, card.King), run(card.Club, card.Ace, card.King),
			}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.board.Stuck(); got != tt.want {
				t.Errorf("Stuck() = %v; want %v", got, tt.want)
			}
		})
	}
}

func TestBoard_Stuck_DoesNotMutate(t *testing.T) {
	b := reportedStuckBoard()
	b.Stuck()
	if !reflect.DeepEqual(b, reportedStuckBoard()) {
		t.Error("expected Stuck to leave the board unchanged")
	}
}

// TestBoard_Stuck_FreshDeals checks Stuck finishes quickly on real
// starting positions, where the search space is largest (full stock, most
// cards face-down). A fresh deal is essentially never provably stuck.
func TestBoard_Stuck_FreshDeals(t *testing.T) {
	for seed := range int64(20) {
		deck := card.NewDeck()
		card.Shuffle(deck, rand.New(rand.NewSource(seed)))
		b, err := solitaire.Deal(deck)
		if err != nil {
			t.Fatal(err)
		}

		start := time.Now()
		stuck := b.Stuck()
		if elapsed := time.Since(start); elapsed > time.Second {
			t.Errorf("seed %d: Stuck took %v", seed, elapsed)
		}
		if stuck {
			t.Logf("seed %d: fresh deal reported stuck", seed)
		}
	}
}

func TestBoard_Clone(t *testing.T) {
	b := reportedStuckBoard()
	c := b.Clone()
	if !reflect.DeepEqual(b, c) {
		t.Fatal("expected Clone to be equal to the original")
	}

	c.Tableau[0].Cards[0] = card.Card{Suit: card.Spade, Rank: card.Ace}
	c.Foundation[card.Spade] = c.Foundation[card.Spade][:1]
	c.Stock = append(c.Stock, card.Card{Suit: card.Spade, Rank: card.Two})
	if !reflect.DeepEqual(b, reportedStuckBoard()) {
		t.Error("expected mutating the Clone to leave the original unchanged")
	}
}
