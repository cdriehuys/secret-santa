package pairings

import (
	"errors"
	"fmt"
	"slices"
)

var (
	ErrNotSolvable = errors.New("the graph is not solvable")

	ErrTooFewNodes = fmt.Errorf("%w: graph must contain at least two nodes", ErrNotSolvable)
	ErrNoPath      = fmt.Errorf("%w: could not find a path that includes every node", ErrNotSolvable)
)

type Graph struct {
	nodes map[string]map[string]struct{}
}

// NewGraphFromExclusions creates a new graph where each node is connected to every other node
// unless the other node is listed in the node's exclusions.
func NewGraphFromExclusions(nodesWithExclusions map[string][]string) *Graph {
	nodes := make(map[string]map[string]struct{}, len(nodesWithExclusions))

	for node, exclusions := range nodesWithExclusions {
		edges := make(map[string]struct{})

		for otherNode := range nodesWithExclusions {
			if otherNode != node && slices.Index(exclusions, otherNode) == -1 {
				edges[otherNode] = struct{}{}
			}
		}

		nodes[node] = edges
	}

	return &Graph{nodes}
}

type Pairing struct {
	From string
	To   string
}

type Random interface {
	Shuffle(n int, swap func(i int, j int))
}

// Pairings generates a random list of pairings such that every node in the graph is both a gifter
// and recipient, and a pairing is only created if there is an edge between the gifter and the
// recipient.
func (g *Graph) Pairings(rand Random) ([]Pairing, error) {
	if len(g.nodes) < 2 {
		return nil, ErrTooFewNodes
	}

	var search func(start string, currentPerson string, visited map[string]struct{}) ([]Pairing, error)

	search = func(start string, currentPerson string, visited map[string]struct{}) ([]Pairing, error) {
		nextCandidates := g.nextCandidates(currentPerson, start, visited)
		shuffle(nextCandidates, rand)

		if len(visited) == len(g.nodes)-1 {
			if _, exists := g.nodes[currentPerson][start]; exists {
				return []Pairing{{From: currentPerson, To: start}}, nil
			}

			// Every node is a gift recipient except for the starting node, but the starting node
			// isn't reachable from the current person, so this solution is not valid.
			return nil, ErrNoPath
		}

		for _, nextPerson := range nextCandidates {
			nextVisited := copyVisitedWithAddition(visited, nextPerson)

			nextSolution, err := search(start, nextPerson, nextVisited)
			if err == nil {
				solution := make([]Pairing, len(nextSolution)+1)
				solution[0] = Pairing{From: currentPerson, To: nextPerson}
				for i, p := range nextSolution {
					solution[i+1] = p
				}

				return solution, nil
			}
		}

		// Every possible path from the current person has been tried, so this solution is invalid.
		return nil, ErrNoPath
	}

	allNodes := make([]string, 0, len(g.nodes))
	for k := range g.nodes {
		allNodes = append(allNodes, k)
	}

	shuffle(allNodes, rand)

	for _, topLevelStart := range allNodes {
		solution, err := search(topLevelStart, topLevelStart, map[string]struct{}{})
		if err == nil {
			return solution, nil
		}
	}

	return nil, ErrNoPath
}

func (g *Graph) nextCandidates(gifter string, start string, visited map[string]struct{}) []string {
	candidates := make([]string, 0, len(g.nodes[gifter]))

	for recipient := range g.nodes[gifter] {
		if recipient != start {
			_, exists := visited[recipient]
			if !exists {
				candidates = append(candidates, recipient)
			}
		}
	}

	return candidates
}

func shuffle(list []string, rand Random) {
	rand.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})
}

func copyVisitedWithAddition(visited map[string]struct{}, addition string) map[string]struct{} {
	copy := make(map[string]struct{}, len(visited)+1)
	for v, s := range visited {
		copy[v] = s
	}

	copy[addition] = struct{}{}

	return copy
}
