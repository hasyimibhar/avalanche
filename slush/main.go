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

	n := 20           // Number of nodes
	m := 20           // Number of rounds
	a := float64(0.7) // Alpha
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

	// Run Slush rounds
	for r := 0; r < m; r++ {
		var wg sync.WaitGroup
		for i := 0; i < n; i++ {
			wg.Add(1)
			go nodes[i].Tick(k, a, &wg)
		}

		wg.Wait()
		log.Println("finished round", r+1)
	}

	// Check final colors
	for i := 0; i < n; i++ {
		fmt.Printf("node %d color=%d\n", i+1, nodes[i].Color())
	}
}
