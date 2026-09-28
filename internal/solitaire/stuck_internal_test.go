package solitaire

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/haruki7049/trump-center/internal/card"
)

func TestStateKey_RoundTrip(t *testing.T) {
	deck := card.NewDeck()
	card.Shuffle(deck, rand.New(rand.NewSource(1)))
	b, err := Deal(deck)
	if err != nil {
		t.Fatal(err)
	}
	// Move a few cards so every section (stock, waste, a pile with
	// several face-up cards, an empty pile) is exercised.
	b.DrawFromStock()
	b.DrawFromStock()
	b.Tableau[1].FaceUp = 2
	b.Tableau[0] = Pile{}

	got := decodeStateKey(b.stateKey(), b.Foundation)
	if !reflect.DeepEqual(got.Tableau, b.Tableau) || !reflect.DeepEqual(got.Stock, b.Stock) || !reflect.DeepEqual(got.Waste, b.Waste) {
		t.Errorf("decodeStateKey(stateKey()) = %+v; want %+v", got, b)
	}
}

// TestStuckWithin_BoundHit checks that running out of search budget
// reports "not stuck", never a false "stuck".
func TestStuckWithin_BoundHit(t *testing.T) {
	// A stuck board whose proof needs more than one state: drawing the
	// unplayable stock card leads to a second state.
	b := &Board{Stock: []card.Card{{Suit: card.Diamond, Rank: card.Five}}}
	b.Tableau[0] = Pile{Cards: []card.Card{{Suit: card.Spade, Rank: card.Ace}, {Suit: card.Heart, Rank: card.Two}}, FaceUp: 1}
	for i := 1; i < TableauPileCount; i++ {
		b.Tableau[i] = Pile{Cards: []card.Card{{Suit: card.Club, Rank: card.King}}, FaceUp: 1}
	}

	if !b.stuckWithin(maxStuckSearchStates) {
		t.Fatal("expected the board to be stuck with the full bound")
	}
	if b.stuckWithin(1) {
		t.Error("expected hitting the bound to report not stuck")
	}
}

// oracleStuck is an independent, deliberately naive reimplementation of
// Stuck, used to cross-check it: it tries every move method with every
// argument that fits the board on a copy (instead of successors'
// Can... checks), doesn't skip column swaps, and keys states on the whole
// board. It returns ok=false if it gives up after maxStates.
func oracleStuck(b *Board, maxStates int) (stuck, ok bool) {
	if b.Won() {
		return false, true
	}
	faceDown, foundation := b.faceDownCount(), b.foundationCount()

	movesFrom := func(cur *Board) []func(n *Board) bool {
		moves := []func(n *Board) bool{
			func(n *Board) bool {
				if len(n.Stock) == 0 && len(n.Waste) == 0 {
					return false
				}
				n.DrawFromStock()
				return true
			},
			func(n *Board) bool { return n.MoveWasteToFoundation() },
		}
		for i := range TableauPileCount {
			moves = append(moves,
				func(n *Board) bool { return n.MoveWasteToTableau(i) },
				func(n *Board) bool { return n.MoveTableauToFoundation(i) },
			)
			for cardIndex := range cur.Tableau[i].Cards {
				for to := range TableauPileCount {
					moves = append(moves, func(n *Board) bool { return n.MoveTableauToTableauRun(i, cardIndex, to) })
				}
			}
		}
		return moves
	}

	key := func(n *Board) string { return fmt.Sprint(n.Tableau, n.Stock, n.Waste) }
	visited := map[string]bool{key(b): true}
	queue := []*Board{b.Clone()}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, move := range movesFrom(cur) {
			n := cur.Clone()
			if !move(n) {
				continue
			}
			if n.faceDownCount() < faceDown || n.foundationCount() > foundation {
				return false, true
			}
			k := key(n)
			if visited[k] {
				continue
			}
			if len(visited) >= maxStates {
				return false, false
			}
			visited[k] = true
			queue = append(queue, n)
		}
	}
	return true, true
}

// randomGames plays seeds random games (picking uniformly among legal
// moves) for up to maxMoves moves each, calling visit on every position.
func randomGames(seeds, maxMoves int, visit func(seed int, b *Board)) {
	for seed := range seeds {
		deck := card.NewDeck()
		r := rand.New(rand.NewSource(int64(seed)))
		card.Shuffle(deck, r)
		b, _ := Deal(deck)
		for range maxMoves {
			visit(seed, b)
			next := b.successors()
			if len(next) == 0 {
				break
			}
			b = next[r.Intn(len(next))]
		}
	}
}

// TestStuck_NeverStuckBeforeProgress plays random games and checks that
// no position was reported stuck if the game later went on to flip a
// card or add one to the foundation from it — which would prove the
// "stuck" wrong.
func TestStuck_NeverStuckBeforeProgress(t *testing.T) {
	type position struct {
		stuck bool
		moves int
	}
	var pending []position
	lastSeed := -1
	var lastFaceDown, lastFoundation int
	checked, stuckSeen := 0, 0

	randomGames(15, 300, func(seed int, b *Board) {
		if seed != lastSeed {
			pending, lastSeed = nil, seed
			lastFaceDown, lastFoundation = b.faceDownCount(), b.foundationCount()
		}
		if b.faceDownCount() < lastFaceDown || b.foundationCount() > lastFoundation {
			for _, p := range pending {
				if p.stuck {
					t.Errorf("seed %d: position at move %d was reported stuck, but progress followed", seed, p.moves)
				}
				checked++
			}
			pending = nil
			lastFaceDown, lastFoundation = b.faceDownCount(), b.foundationCount()
		}

		s := b.Stuck()
		if s {
			stuckSeen++
		}
		pending = append(pending, position{stuck: s, moves: len(pending)})
	})

	if checked == 0 || stuckSeen == 0 {
		t.Fatalf("test didn't exercise both cases: %d positions checked, %d stuck seen", checked, stuckSeen)
	}
}

// TestStuck_MatchesOracle cross-checks Stuck against oracleStuck on
// positions from random games: the first two each game reports stuck
// (a random game keeps playing once stuck, so later ones are near
// repeats), plus every tenth other one. The naive oracle is too slow to
// check them all on every test run.
func TestStuck_MatchesOracle(t *testing.T) {
	compared, stuckCompared, n := 0, 0, 0
	stuckPerSeed := map[int]int{}
	randomGames(10, 300, func(seed int, b *Board) {
		got := b.Stuck()
		n++
		if got {
			stuckPerSeed[seed]++
			if stuckPerSeed[seed] > 2 {
				return
			}
		} else if n%10 != 0 {
			return
		}

		want, ok := oracleStuck(b, 2_000)
		if !ok {
			return
		}
		if got != want {
			t.Errorf("seed %d: Stuck() = %v; oracle says %v for %+v", seed, got, want, b)
		}
		compared++
		if want {
			stuckCompared++
		}
	})

	if stuckCompared == 0 {
		t.Fatalf("no stuck positions were compared (out of %d)", compared)
	}
	t.Logf("compared %d positions, %d of them stuck", compared, stuckCompared)
}
