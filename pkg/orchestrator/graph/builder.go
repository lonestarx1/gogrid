package graph

import (
	"fmt"
	"sort"

	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/trace"
)

// Builder provides a fluent API for constructing graphs.
type Builder struct {
	name  string
	nodes map[string]*Node
	edges []Edge
	opts  []Option
	err   error
}

// NewBuilder creates a graph builder with the given name.
func NewBuilder(name string) *Builder {
	return &Builder{
		name:  name,
		nodes: make(map[string]*Node),
	}
}

// AddNode adds a named node wrapping an agent.
// Returns the builder for chaining.
func (b *Builder) AddNode(name string, a *agent.Agent) *Builder {
	if b.err != nil {
		return b
	}
	if _, exists := b.nodes[name]; exists {
		b.err = fmt.Errorf("graph: duplicate node %q", name)
		return b
	}
	b.nodes[name] = &Node{name: name, agent: a}
	return b
}

// AddEdge adds a directed edge from one node to another.
// An optional EdgeCondition controls whether the edge is followed.
// If no condition is provided, the edge is always followed.
func (b *Builder) AddEdge(from, to string, conditions ...EdgeCondition) *Builder {
	if b.err != nil {
		return b
	}
	var cond EdgeCondition
	if len(conditions) > 0 {
		cond = conditions[0]
	}
	b.edges = append(b.edges, Edge{From: from, To: to, Condition: cond})
	return b
}

// Options adds functional options that will be applied to the graph.
func (b *Builder) Options(opts ...Option) *Builder {
	b.opts = append(b.opts, opts...)
	return b
}

// Build validates the graph and returns it.
// Returns an error if:
//   - No nodes are defined
//   - An edge references a nonexistent node
//   - Duplicate node names exist (caught during AddNode)
func (b *Builder) Build() (*Graph, error) {
	if b.err != nil {
		return nil, b.err
	}
	if len(b.nodes) == 0 {
		return nil, fmt.Errorf("graph: no nodes defined")
	}

	// Validate edges reference existing nodes.
	for _, e := range b.edges {
		if _, ok := b.nodes[e.From]; !ok {
			return nil, fmt.Errorf("graph: edge references unknown node %q", e.From)
		}
		if _, ok := b.nodes[e.To]; !ok {
			return nil, fmt.Errorf("graph: edge references unknown node %q", e.To)
		}
	}

	// Build adjacency structures.
	outgoing := make(map[string][]Edge)
	incomingSet := make(map[string]map[string]bool)

	for name := range b.nodes {
		outgoing[name] = nil
		incomingSet[name] = make(map[string]bool)
	}

	for _, e := range b.edges {
		outgoing[e.From] = append(outgoing[e.From], e)
		incomingSet[e.To][e.From] = true
	}

	// Convert incoming sets to sorted slices.
	incoming := make(map[string][]string)
	for name, sources := range incomingSet {
		sl := make([]string, 0, len(sources))
		for src := range sources {
			sl = append(sl, src)
		}
		sort.Strings(sl)
		incoming[name] = sl
	}

	// Find start nodes (no incoming edges).
	var starts []string
	for name := range b.nodes {
		if len(incoming[name]) == 0 {
			starts = append(starts, name)
		}
	}
	sort.Strings(starts)

	// Find end nodes (no outgoing edges).
	var ends []string
	for name := range b.nodes {
		if len(outgoing[name]) == 0 {
			ends = append(ends, name)
		}
	}
	sort.Strings(ends)

	g := &Graph{
		name:     b.name,
		nodes:    b.nodes,
		edges:    b.edges,
		outgoing: outgoing,
		incoming: incoming,
		starts:   starts,
		ends:     ends,
		tracer:   trace.Noop{},
	}

	for _, opt := range b.opts {
		opt(g)
	}

	return g, nil
}
