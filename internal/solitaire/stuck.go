package solitaire

import "github.com/haruki7049/trump-center/internal/card"

// maxStuckSearchStates bounds how many distinct board states Stuck
// explores before giving up. Giving up always reports "not stuck", so the
// bound can only ever cost a missed "No moves left", never a false one.
// In random play, proving a position stuck took up to ~120,000 states
// (median ~30); this bound covers that with room to spare, at up to about
// two seconds and a few tens of MB, which is why callers should run
// Stuck off the game loop (see Clone).
const maxStuckSearchStates = 300_000

// Stuck reports whether the game can no longer be won from this position:
// no sequence of legal moves (including drawing from and recycling the
// stock) ever flips a face-down tableau card or adds a card to the
// foundation. Since winning requires filling the foundation, a position
// where neither can ever happen again is lost; the converse doesn't hold
// (progress may still be possible in a lost game), so Stuck only reports
// the positions it can prove are lost.
//
// It explores every reachable position with a bounded breadth-first
// search, and returns false for a won game, or if the search hits
// maxStuckSearchStates before finishing.
func (b *Board) Stuck() bool {
	return b.stuckWithin(maxStuckSearchStates)
}

// stuckWithin is Stuck with the search bound as a parameter, so tests can
// exercise what happens when the bound is hit.
func (b *Board) stuckWithin(maxStates int) bool {
	if b.Won() {
		return false
	}

	start := b.Clone()
	faceDown := start.faceDownCount()
	foundation := start.foundationCount()

	// Queued states are kept as their compact stateKey (tens of bytes)
	// rather than as *Board (about a kilobyte), and decoded when visited.
	startKey := start.stateKey()
	visited := map[string]struct{}{startKey: {}}
	queue := []string{startKey}
	for len(queue) > 0 {
		cur := decodeStateKey(queue[0], start.Foundation)
		queue = queue[1:]

		for _, next := range cur.successors() {
			if next.faceDownCount() < faceDown || next.foundationCount() > foundation {
				return false
			}

			key := next.stateKey()
			if _, ok := visited[key]; ok {
				continue
			}
			if len(visited) >= maxStates {
				return false
			}
			visited[key] = struct{}{}
			queue = append(queue, key)
		}
	}
	return true
}

// successors returns a copy of b for each legal move from it, with that
// move applied.
func (b *Board) successors() []*Board {
	var next []*Board
	apply := func(move func(n *Board)) {
		n := b.Clone()
		move(n)
		next = append(next, n)
	}

	if len(b.Stock) > 0 || len(b.Waste) > 0 {
		apply(func(n *Board) { n.DrawFromStock() })
	}

	if c, ok := b.WasteTop(); ok {
		if b.CanMoveToFoundation(c) {
			apply(func(n *Board) { n.MoveWasteToFoundation() })
		}
		for to := range TableauPileCount {
			if b.CanMoveToTableau(c, to) {
				apply(func(n *Board) { n.MoveWasteToTableau(to) })
			}
		}
	}

	for from, pile := range b.Tableau {
		if c, ok := pile.Top(); ok && pile.FaceUp > 0 && b.CanMoveToFoundation(c) {
			apply(func(n *Board) { n.MoveTableauToFoundation(from) })
		}

		faceDownCount := len(pile.Cards) - pile.FaceUp
		for cardIndex := faceDownCount; cardIndex < len(pile.Cards); cardIndex++ {
			for to := range TableauPileCount {
				// Moving a whole pile (nothing beneath it) into an empty
				// pile only swaps two columns, which can't change what's
				// reachable; skipping it keeps the search small.
				if cardIndex == 0 && len(b.Tableau[to].Cards) == 0 {
					continue
				}
				if b.CanMoveTableauRunToTableau(from, cardIndex, to) {
					apply(func(n *Board) { n.MoveTableauToTableauRun(from, cardIndex, to) })
				}
			}
		}
	}

	return next
}

// Clone returns a deep copy of b. Since Stuck can take a noticeable
// fraction of a second, a UI can Clone the board and run Stuck on the
// copy in the background while the player keeps playing on the original.
func (b *Board) Clone() *Board {
	n := &Board{
		Stock: append([]card.Card(nil), b.Stock...),
		Waste: append([]card.Card(nil), b.Waste...),
	}
	for i, pile := range b.Foundation {
		n.Foundation[i] = append([]card.Card(nil), pile...)
	}
	for i, pile := range b.Tableau {
		n.Tableau[i] = Pile{Cards: append([]card.Card(nil), pile.Cards...), FaceUp: pile.FaceUp}
	}
	return n
}

func (b *Board) faceDownCount() int {
	n := 0
	for _, pile := range b.Tableau {
		n += len(pile.Cards) - pile.FaceUp
	}
	return n
}

func (b *Board) foundationCount() int {
	n := 0
	for _, pile := range b.Foundation {
		n += len(pile)
	}
	return n
}

// stateKeySep separates piles in a stateKey. It can't collide with an
// encoded card (at most 0x3d) or a FaceUp count (at most 19).
const stateKeySep = 0xff

// stateKey encodes the parts of b that can change without progress (see
// Stuck) — the tableau, stock, and waste — one byte per card or FaceUp
// count. The foundation is left out, since Stuck stops as soon as it
// changes; decodeStateKey takes it back as an argument.
func (b *Board) stateKey() string {
	encode := func(c card.Card) byte { return byte(c.Suit)<<4 | byte(c.Rank) }

	key := make([]byte, 0, 80)
	for _, pile := range b.Tableau {
		key = append(key, byte(pile.FaceUp))
		for _, c := range pile.Cards {
			key = append(key, encode(c))
		}
		key = append(key, stateKeySep)
	}
	for _, c := range b.Stock {
		key = append(key, encode(c))
	}
	key = append(key, stateKeySep)
	for _, c := range b.Waste {
		key = append(key, encode(c))
	}
	return string(key)
}

// decodeStateKey rebuilds the Board a stateKey was made from, given its
// foundation. The foundation slices are shared, not copied; that's safe
// because successors only ever mutates Clones.
func decodeStateKey(key string, foundation [4][]card.Card) *Board {
	decode := func(x byte) card.Card { return card.Card{Suit: card.Suit(x >> 4), Rank: card.Rank(x & 0x0f)} }

	b := &Board{Foundation: foundation}
	i := 0
	for p := range TableauPileCount {
		faceUp := int(key[i])
		i++
		var cards []card.Card
		for ; key[i] != stateKeySep; i++ {
			cards = append(cards, decode(key[i]))
		}
		i++
		b.Tableau[p] = Pile{Cards: cards, FaceUp: faceUp}
	}
	for ; key[i] != stateKeySep; i++ {
		b.Stock = append(b.Stock, decode(key[i]))
	}
	for i++; i < len(key); i++ {
		b.Waste = append(b.Waste, decode(key[i]))
	}
	return b
}
