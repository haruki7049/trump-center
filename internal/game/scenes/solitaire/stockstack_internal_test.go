package solitaire

import (
	"testing"

	"github.com/haruki7049/trump-center/internal/card"
	"github.com/haruki7049/trump-center/internal/solitaire"
)

func TestSolitaireScene_SyncComponents_StockStackedLayers(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{
		Stock: []card.Card{
			{Suit: card.Spade, Rank: card.Ace},
			{Suit: card.Spade, Rank: card.Two},
			{Suit: card.Spade, Rank: card.Three},
			{Suit: card.Spade, Rank: card.Four},
			{Suit: card.Spade, Rank: card.Five},
			{Suit: card.Spade, Rank: card.Six},
		},
	})
	s.syncComponents()

	// The front (clickable) layer must stay exactly at the pile's
	// original fixed position, regardless of how many layers are drawn
	// behind it, so tryStartDrag/hit-testing bounds don't need to change.
	front, ok := s.stockPile.TopBounds()
	if !ok {
		t.Fatal("expected a front card")
	}
	if front.Min.X != stockOriginX || front.Min.Y != stockOriginY {
		t.Errorf("front card at (%d, %d); want (%d, %d)", front.Min.X, front.Min.Y, stockOriginX, stockOriginY)
	}

	// With more cards than stockMaxLayers, exactly stockMaxLayers should
	// be drawn, each offset further up-left than the last, back to front.
	for i := range stockMaxLayers {
		b, ok := s.stockPile.CardBoundsAt(i)
		if !ok {
			t.Fatalf("expected layer %d to be drawn", i)
		}
		depth := stockMaxLayers - 1 - i
		wantX := stockOriginX - depth*stockLayerOffset
		wantY := stockOriginY - depth*stockLayerOffset
		if b.Min.X != wantX || b.Min.Y != wantY {
			t.Errorf("layer %d at (%d, %d); want (%d, %d)", i, b.Min.X, b.Min.Y, wantX, wantY)
		}
	}
	if _, ok := s.stockPile.CardBoundsAt(stockMaxLayers); ok {
		t.Errorf("expected at most %d layers to be drawn", stockMaxLayers)
	}
}

func TestSolitaireScene_SyncComponents_StockStackedLayers_FewerCardsThanMax(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{
		Stock: []card.Card{
			{Suit: card.Heart, Rank: card.Ace},
			{Suit: card.Heart, Rank: card.Two},
		},
	})
	s.syncComponents()

	front, ok := s.stockPile.TopBounds()
	if !ok {
		t.Fatal("expected a front card")
	}
	if front.Min.X != stockOriginX || front.Min.Y != stockOriginY {
		t.Errorf("front card at (%d, %d); want (%d, %d)", front.Min.X, front.Min.Y, stockOriginX, stockOriginY)
	}
	if _, ok := s.stockPile.CardBoundsAt(2); ok {
		t.Error("expected only 2 layers with 2 cards in the stock")
	}
}

func TestSolitaireScene_SyncComponents_StockEmpty_NoLayers(t *testing.T) {
	s := newTestScene(t)
	s.setBoard(&solitaire.Board{})
	s.syncComponents()

	if _, ok := s.stockPile.TopBounds(); ok {
		t.Error("expected no stock layers when the stock is empty")
	}
}
