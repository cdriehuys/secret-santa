package pairings_test

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"testing"

	"github.com/cdriehuys/secret-santa/internal/pairings"
)

func TestGraph_Pairings_visitsAll(t *testing.T) {
	testCases := []struct {
		name  string
		nodes map[string][]string
	}{
		{
			name: "two people no exclusions",
			nodes: map[string][]string{
				"Jane": nil,
				"Bob":  nil,
			},
		},
		{
			name: "five people no exclusions",
			nodes: map[string][]string{
				"Bob":     nil,
				"Sally":   nil,
				"Charlie": nil,
				"Kim":     nil,
				"Andy":    nil,
			},
		},
		{
			name: "basic exclusion",
			nodes: map[string][]string{
				"Ross":   {"Monica"},
				"Monica": {"Rachel"},
				"Rachel": {"Ross"},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			graph := pairings.NewGraphFromExclusions(tt.nodes)
			pairs, err := graph.Pairings(fixedRandom())

			if err != nil {
				t.Fatalf("Unable to generate pairings: %v", err)
			}

			var gifters []string
			var recipients []string

			for _, p := range pairs {
				gifter := p.From
				recipient := p.To

				if slices.Index(gifters, gifter) != -1 {
					t.Errorf("%s is giving more than one gift", gifter)
				}

				gifters = append(gifters, gifter)

				if slices.Index(recipients, recipient) != -1 {
					t.Errorf("%s is receiving more than one gift", recipient)
				}

				recipients = append(recipients, recipient)
			}

			for node := range tt.nodes {
				if slices.Index(gifters, node) == -1 {
					t.Errorf("%s is not giving a gift", node)
				}

				if slices.Index(recipients, node) == -1 {
					t.Errorf("%s is not receiving a gift", node)
				}
			}
		})
	}
}

func TestGraph_Pairings_errorCases(t *testing.T) {
	testCases := []struct {
		name      string
		nodes     map[string][]string
		wantError error
	}{
		{
			name:      "empty graph",
			nodes:     map[string][]string{},
			wantError: pairings.ErrTooFewNodes,
		},
		{
			name: "one node",
			nodes: map[string][]string{
				"Aristotle": nil,
			},
			wantError: pairings.ErrTooFewNodes,
		},
		{
			name: "simple unsolvable",
			nodes: map[string][]string{
				"Percy":  {"Edward"},
				"Edward": nil,
			},
			wantError: pairings.ErrNoPath,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			graph := pairings.NewGraphFromExclusions(tt.nodes)
			_, err := graph.Pairings(fixedRandom())

			if err == nil {
				t.Error("No error was returned")
			}

			if !errors.Is(err, tt.wantError) {
				t.Errorf("Expected error %#v, received %#v", tt.wantError, err)
			}
		})
	}
}

func BenchmarkGraph_Pairings_NoExclusions(b *testing.B) {
	nodes := make(map[string][]string)
	for i := range 10 {
		nodes[fmt.Sprintf("N%d", i)] = nil
	}

	graph := pairings.NewGraphFromExclusions(nodes)
	rand := fixedRandom()

	for b.Loop() {
		graph.Pairings(rand)
	}
}

func BenchmarkGraph_Pairings_PartneredExclusions(b *testing.B) {
	nodes := make(map[string][]string)
	for i := range 5 {
		n1 := fmt.Sprintf("N%d", 2*i)
		n2 := fmt.Sprintf("N%d", 2*i+1)
		nodes[n1] = []string{n2}
		nodes[n2] = []string{n1}
	}

	graph := pairings.NewGraphFromExclusions(nodes)
	rand := fixedRandom()

	for b.Loop() {
		graph.Pairings(rand)
	}
}

// fixedRandom returns a random instance with a fixed seed so that any test failures can be
// consistently reproduced.
func fixedRandom() *rand.Rand {
	return rand.New(rand.NewSource(42))
}
