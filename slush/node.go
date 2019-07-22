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
	mtx   sync.Mutex
	color Color
	Peers []*Node
}

func (n *Node) SetColor(col Color) {
	n.mtx.Lock()
	n.color = col
	n.mtx.Unlock()
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

// Tick runs a single round of Slush.
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

	// Flip color if count[Color] >= alpha*k
	n.mtx.Lock()
	if n.color == ColorBlue && float64(count[ColorRed]) >= alpha*float64(k) {
		n.color = ColorRed
	} else if n.color == ColorRed && float64(count[ColorBlue]) >= alpha*float64(k) {
		n.color = ColorBlue
	}
	n.mtx.Unlock()
}
