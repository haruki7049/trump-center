package card_test

import (
	"math/rand"
	"testing"

	"github.com/haruki7049/trump-center/internal/card"
)

func TestSuit_String(t *testing.T) {
	tests := []struct {
		suit card.Suit
		want string
	}{
		{card.Spade, "spade"},
		{card.Heart, "heart"},
		{card.Diamond, "diamond"},
		{card.Club, "club"},
		{card.Suit(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.suit.String(); got != tt.want {
			t.Errorf("Suit(%d).String() = %q; want %q", tt.suit, got, tt.want)
		}
	}
}

func TestSuit_Red(t *testing.T) {
	tests := []struct {
		suit card.Suit
		want bool
	}{
		{card.Spade, false},
		{card.Club, false},
		{card.Heart, true},
		{card.Diamond, true},
	}

	for _, tt := range tests {
		if got := tt.suit.Red(); got != tt.want {
			t.Errorf("Suit(%v).Red() = %v; want %v", tt.suit, got, tt.want)
		}
	}
}

func TestCard_AssetPath(t *testing.T) {
	c := card.Card{Suit: card.Spade, Rank: card.Ace}
	want := "cards/spade/01.png"
	if got := c.AssetPath(); got != want {
		t.Errorf("AssetPath() = %q; want %q", got, want)
	}

	c = card.Card{Suit: card.Heart, Rank: card.King}
	want = "cards/heart/13.png"
	if got := c.AssetPath(); got != want {
		t.Errorf("AssetPath() = %q; want %q", got, want)
	}
}

func TestJoker_AssetPath(t *testing.T) {
	want := "cards/joker.png"
	if got := (card.Joker{}).AssetPath(); got != want {
		t.Errorf("AssetPath() = %q; want %q", got, want)
	}
}

func TestNewDeck(t *testing.T) {
	deck := card.NewDeck()
	if len(deck) != 52 {
		t.Fatalf("expected 52 cards, got %d", len(deck))
	}

	seen := make(map[card.Card]bool)
	for _, c := range deck {
		if seen[c] {
			t.Fatalf("duplicate card found: %+v", c)
		}
		seen[c] = true

		if c.Rank < card.Ace || c.Rank > card.King {
			t.Errorf("rank out of range: %+v", c)
		}
	}
}

func TestShuffle(t *testing.T) {
	deck := card.NewDeck()
	original := make([]card.Card, len(deck))
	copy(original, deck)

	card.Shuffle(deck, rand.New(rand.NewSource(1)))

	if len(deck) != len(original) {
		t.Fatalf("shuffle changed deck length: got %d, want %d", len(deck), len(original))
	}

	same := true
	for i := range deck {
		if deck[i] != original[i] {
			same = false
			break
		}
	}
	if same {
		t.Errorf("expected shuffle to change card order")
	}

	counts := make(map[card.Card]int)
	for _, c := range deck {
		counts[c]++
	}
	for _, c := range original {
		if counts[c] != 1 {
			t.Errorf("card %+v appears %d times after shuffle; want 1", c, counts[c])
		}
	}
}
