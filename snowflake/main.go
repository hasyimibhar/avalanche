package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	n := 50           // Number of nodes
	a := float64(0.7) // Alpha
	b := uint(4)      // Beta
	k := 5            // k sample size

	// Initialize node initial state and peers
	nodes := []*Node{}
	for i := 0; i < n; i++ {
		nodes = append(nodes, &Node{})
	}

	for i := 0; i < n; i++ {
		peers := []*Node{}
		for j := 0; j < (n / 2); j++ {
			peers = append(peers, nodes[(i+j)%(n/2)])
		}
		nodes[i].Peers = peers
		nodes[i].SetColor(Color((i % 2) + 1))
	}

	var decidedCount int
	log.Println("waiting for nodes to decide...")

	// Run Snowflake rounds until no nodes are undecided
	for decidedCount < n {
		var wg sync.WaitGroup
		undecidedNodes := []*Node{}
		for i := 0; i < n; i++ {
			if nodes[i].Decided(b) {
				continue
			}

			wg.Add(1)
			go nodes[i].Tick(k, a, &wg)
			undecidedNodes = append(undecidedNodes, nodes[i])
		}

		wg.Wait()

		for _, node := range undecidedNodes {
			if node.Decided(b) {
				decidedCount++
			}
		}
	}

	// Check final colors
	for i := 0; i < n; i++ {
		fmt.Printf("node %d color=%d\n", i+1, nodes[i].Color())
	}
}
