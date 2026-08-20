package domain

import (
	"fmt"
	"sort"
)

type Node struct {
	ID           string
	Name         string
	Dependencies []string
}

type Graph struct {
	nodes map[string]Node
}

func NewGraph() *Graph {
	return &Graph{nodes: make(map[string]Node)}
}

func (g *Graph) AddNode(node Node) {
	g.nodes[node.ID] = node
}

func (g *Graph) Node(id string) (Node, bool) {
	n, ok := g.nodes[id]
	return n, ok
}

func (g *Graph) Nodes() []Node {
	out := make([]Node, 0, len(g.nodes))
	ids := make([]string, 0, len(g.nodes))
	for id := range g.nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		out = append(out, g.nodes[id])
	}
	return out
}

func (g *Graph) Validate() error {
	visited := make(map[string]int)
	var visit func(string, []string) error
	visit = func(id string, stack []string) error {
		if visited[id] == 2 {
			return nil
		}
		if visited[id] == 1 {
			return &CycleError{Path: append(append([]string{}, stack...), id)}
		}
		node, ok := g.nodes[id]
		if !ok {
			return &MissingNodeError{ID: id}
		}
		visited[id] = 1
		for _, dep := range node.Dependencies {
			if err := visit(dep, append(stack, id)); err != nil {
				return err
			}
		}
		visited[id] = 2
		return nil
	}
	for id := range g.nodes {
		if err := visit(id, nil); err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph) TopologicalLevels() ([][]string, error) {
	if err := g.Validate(); err != nil {
		return nil, err
	}
	indegree := make(map[string]int)
	children := make(map[string][]string)
	for id := range g.nodes {
		indegree[id] = 0
	}
	for _, node := range g.nodes {
		for _, dep := range node.Dependencies {
			indegree[node.ID]++
			children[dep] = append(children[dep], node.ID)
		}
	}
	var ready []string
	for id, deg := range indegree {
		if deg == 0 {
			ready = append(ready, id)
		}
	}
	sort.Strings(ready)
	var levels [][]string
	visitedCount := 0
	for len(ready) > 0 {
		levels = append(levels, ready)
		visitedCount += len(ready)
		var next []string
		for _, id := range ready {
			for _, child := range children[id] {
				indegree[child]--
				if indegree[child] == 0 {
					next = append(next, child)
				}
			}
		}
		sort.Strings(next)
		ready = next
	}
	if visitedCount != len(g.nodes) {
		return nil, fmt.Errorf("incomplete topological order")
	}
	return levels, nil
}

type CycleError struct {
	Path []string
}

func (e *CycleError) Error() string {
	return fmt.Sprintf("dependency cycle: %v", e.Path)
}

type MissingNodeError struct {
	ID string
}

func (e *MissingNodeError) Error() string {
	return fmt.Sprintf("dependency points to missing resource %q", e.ID)
}
