package main

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type Color int

// Set of possible colors
const (
	ColorUncolored Color = iota
	ColorRed
	ColorBlue
)

// Node represents a single Slush node.
type Node struct {
	mtx              sync.Mutex
	color, prevColor Color
	confidence       map[Color]uint
	counter          uint
	Peers            []*Node
}

func NewNode(col Color) *Node {
	return &Node{
		Peers:      []*Node{},
		color:      col,
		prevColor:  col,
		confidence: map[Color]uint{},
	}
}

func (n *Node) Color() Color {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	return n.color
}

// Query accepts a color. If node is uncolored, it adopts the color.
// Otherwise, it just responds with its current color.
func (n *Node) Query(ctx context.Context, c Color) (Color, error) {
	n.mtx.Lock()
	defer n.mtx.Unlock()

	select {
	// Simulate network latency
	case <-time.After(time.Millisecond*5 + time.Millisecond*time.Duration(rand.Intn(10))):
		if n.color == ColorUncolored {
			n.color = c
		}
		return n.color, nil

	case <-ctx.Done():
		return n.color, ctx.Err()
	}
}

func (n *Node) Decided(beta uint) bool {
	n.mtx.Lock()
	defer n.mtx.Unlock()

	return n.counter > beta
}

// Tick runs a single round of Snowball.
func (n *Node) Tick(k int, alpha float64, cwg *sync.WaitGroup) {
	defer cwg.Done()

	// If node is uncolored, do nothing
	n.mtx.Lock()
	if n.color == ColorUncolored {
		n.mtx.Unlock()
		return
	}
	n.mtx.Unlock()

	// Select a random at most k samples from the node's peers
	sample := make([]*Node, len(n.Peers))
	for i, peer := range n.Peers {
		sample[i] = peer
	}
	rand.Shuffle(len(sample), func(i, j int) { sample[i], sample[j] = sample[j], sample[i] })

	if k < len(sample) {
		sample = sample[:k]
	}

	var wg sync.WaitGroup
	responses := []Color{}
	var mtx sync.Mutex

	// Query the sampled peers for its color
	for _, peer := range sample {
		wg.Add(1)
		go func(p *Node) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
			defer cancel()

			resp, err := p.Query(ctx, n.color)
			if err != nil {
				return
			}

			mtx.Lock()
			responses = append(responses, resp)
			mtx.Unlock()
		}(peer)
	}

	wg.Wait()
	count := map[Color]int{}
	for _, col := range responses {
		count[col]++
	}

	n.mtx.Lock()
	defer n.mtx.Unlock()

	// Flip color if count[Color] >= alpha*k
	for i := 1; i <= 2; i++ {
		if float64(count[Color(i)]) >= alpha*float64(k) {
			n.confidence[Color(i)]++
			if n.confidence[Color(i)] > n.confidence[n.color] {
				n.color = Color(i)
			} else if Color(i) != n.prevColor {
				n.prevColor = Color(i)
				n.counter = 0
			} else {
				n.counter++
			}
			break
		}
	}
}
