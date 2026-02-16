package graph

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lonestarx1/gogrid/internal/id"
	"github.com/lonestarx1/gogrid/pkg/agent"
	"github.com/lonestarx1/gogrid/pkg/llm"
	"github.com/lonestarx1/gogrid/pkg/trace"
)

// EdgeCondition is a function that determines whether an edge should
// be followed based on the source node's result.
type EdgeCondition func(ctx context.Context, result *NodeResult) (bool, error)

// When wraps a condition function for use in edge definitions.
func When(fn func(output string) bool) EdgeCondition {
	return func(_ context.Context, r *NodeResult) (bool, error) {
		return fn(r.Output), nil
	}
}

// Always returns an EdgeCondition that always evaluates to true.
func Always() EdgeCondition {
	return func(_ context.Context, _ *NodeResult) (bool, error) {
		return true, nil
	}
}

// Node wraps an agent as a graph vertex.
type Node struct {
	name  string
	agent *agent.Agent
}

// Edge connects two nodes with an optional condition.
type Edge struct {
	From      string
	To        string
	Condition EdgeCondition // nil means always follow
}

// NodeResult holds the output of a single node execution.
type NodeResult struct {
	// Name is the node name.
	Name string
	// Output is the agent's response content.
	Output string
	// AgentResult is the full agent result.
	AgentResult *agent.Result
	// Iteration is which iteration of this node (1-based, >1 for loops).
	Iteration int
}

// Config controls graph execution behavior.
type Config struct {
	// MaxIterations limits how many times any node can be visited
	// (loop guard). 0 defaults to 10.
	MaxIterations int
	// Timeout is the maximum wall-clock duration for the entire graph.
	// Zero means no timeout.
	Timeout time.Duration
	// CostBudget is the maximum total cost in USD. Zero means no limit.
	CostBudget float64
}

// Result is returned by Graph.Run.
type Result struct {
	// RunID uniquely identifies this execution.
	RunID string
	// Output is the final output from the terminal node(s).
	// If multiple terminal nodes, their outputs are joined.
	Output string
	// NodeResults maps node name to its results (may have multiple
	// entries if a node was visited more than once due to loops).
	NodeResults map[string][]*NodeResult
	// TotalCost is the aggregate cost in USD.
	TotalCost float64
	// TotalUsage is the aggregate token usage.
	TotalUsage llm.Usage
}

// Graph orchestrates agents as nodes in a directed graph with
// conditional edges, parallel fan-out, fan-in merging, and loops.
type Graph struct {
	name   string
	nodes  map[string]*Node
	edges  []Edge
	config Config
	tracer trace.Tracer

	// Derived from edges during Build.
	outgoing map[string][]Edge // node -> outgoing edges
	incoming map[string][]string // node -> list of source node names
	starts   []string           // nodes with no incoming edges
	ends     []string           // nodes with no outgoing edges
}

// Option is a functional option for configuring a Graph.
type Option func(*Graph)

// WithConfig sets the graph's execution configuration.
func WithConfig(c Config) Option {
	return func(g *Graph) {
		g.config = c
	}
}

// WithTracer sets the tracer for observability.
func WithTracer(t trace.Tracer) Option {
	return func(g *Graph) {
		g.tracer = t
	}
}

// Name returns the graph's name.
func (g *Graph) Name() string { return g.name }

// Nodes returns a sorted list of node names.
func (g *Graph) Nodes() []string {
	names := make([]string, 0, len(g.nodes))
	for name := range g.nodes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Edges returns a copy of the graph's edges.
func (g *Graph) Edges() []Edge {
	cp := make([]Edge, len(g.edges))
	copy(cp, g.edges)
	return cp
}

// DOT exports the graph in Graphviz DOT format for visualization.
func (g *Graph) DOT() string {
	var b strings.Builder
	b.WriteString("digraph ")
	b.WriteString(strconv.Quote(g.name))
	b.WriteString(" {\n")
	b.WriteString("  rankdir=LR;\n")
	b.WriteString("  node [shape=box, style=rounded];\n")
	for _, name := range g.Nodes() {
		b.WriteString("  ")
		b.WriteString(strconv.Quote(name))
		b.WriteString(";\n")
	}
	for _, e := range g.edges {
		b.WriteString("  ")
		b.WriteString(strconv.Quote(e.From))
		b.WriteString(" -> ")
		b.WriteString(strconv.Quote(e.To))
		if e.Condition != nil {
			b.WriteString(" [style=dashed]")
		}
		b.WriteString(";\n")
	}
	b.WriteString("}\n")
	return b.String()
}

// Run executes the graph starting from root nodes (no incoming edges).
//
// Execution model:
//  1. Find start nodes (no incoming edges).
//  2. Run all ready nodes concurrently.
//  3. When a node completes, evaluate outgoing edge conditions.
//  4. Mark successor nodes as ready when all their incoming edges
//     have been satisfied.
//  5. Repeat until no more nodes are ready or limits are hit.
func (g *Graph) Run(ctx context.Context, input string) (*Result, error) {
	if len(g.nodes) == 0 {
		return nil, errors.New("graph: no nodes defined")
	}
	if len(g.starts) == 0 {
		return nil, errors.New("graph: no start nodes (all nodes have incoming edges)")
	}

	// Apply timeout.
	if g.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.config.Timeout)
		defer cancel()
	}

	runID := id.New()

	ctx, runSpan := g.tracer.StartSpan(ctx, "graph.run")
	runSpan.SetAttribute("graph.name", g.name)
	runSpan.SetAttribute("graph.run_id", runID)
	runSpan.SetAttribute("graph.nodes", strconv.Itoa(len(g.nodes)))
	runSpan.SetAttribute("graph.edges", strconv.Itoa(len(g.edges)))
	defer g.tracer.EndSpan(runSpan)

	maxIter := g.config.MaxIterations
	if maxIter <= 0 {
		maxIter = 10
	}

	// Execution state.
	var mu sync.Mutex
	nodeResults := make(map[string][]*NodeResult)
	iterations := make(map[string]int)     // visit count per node
	nodeOutputs := make(map[string]string) // latest output per node
	var totalCost float64
	var totalUsage llm.Usage

	// Initialize: start nodes are ready with the initial input.
	ready := make([]string, len(g.starts))
	copy(ready, g.starts)

	for len(ready) > 0 {
		if err := ctx.Err(); err != nil {
			runSpan.SetError(err)
			return nil, fmt.Errorf("graph: %w", err)
		}

		mu.Lock()
		if g.config.CostBudget > 0 && totalCost >= g.config.CostBudget {
			mu.Unlock()
			runSpan.SetAttribute("graph.stopped_reason", "cost_budget")
			break
		}
		mu.Unlock()

		// Track which nodes completed in THIS wave only.
		waveCompleted := make(map[string]bool)

		// Run all ready nodes concurrently.
		type nodeResponse struct {
			name   string
			result *NodeResult
			err    error
		}
		respCh := make(chan nodeResponse, len(ready))

		for _, nodeName := range ready {
			mu.Lock()
			iterations[nodeName]++
			iter := iterations[nodeName]
			if iter > maxIter {
				mu.Unlock()
				runSpan.SetAttribute("graph.stopped_reason", "max_iterations_"+nodeName)
				respCh <- nodeResponse{
					name: nodeName,
					err:  fmt.Errorf("node %q exceeded max iterations (%d)", nodeName, maxIter),
				}
				continue
			}

			// Determine input: merge outputs from all incoming nodes,
			// or use the initial input if this is a start node.
			nodeInput := g.buildNodeInput(nodeName, input, nodeOutputs)
			mu.Unlock()

			go func(name string, inp string, iteration int) {
				nr, err := g.runNode(ctx, name, inp, iteration)
				respCh <- nodeResponse{name: name, result: nr, err: err}
			}(nodeName, nodeInput, iter)
		}

		// Collect results.
		for range ready {
			resp := <-respCh

			mu.Lock()
			if resp.err != nil {
				// Propagate context errors (timeout, cancellation).
				if ctx.Err() != nil {
					mu.Unlock()
					runSpan.SetError(ctx.Err())
					return nil, fmt.Errorf("graph: %w", ctx.Err())
				}
				runSpan.SetAttribute("graph.node."+resp.name+".error", resp.err.Error())
				mu.Unlock()
				continue
			}

			nr := resp.result
			nodeResults[nr.Name] = append(nodeResults[nr.Name], nr)
			nodeOutputs[nr.Name] = nr.Output
			waveCompleted[nr.Name] = true

			if nr.AgentResult != nil {
				totalCost += nr.AgentResult.Cost
				totalUsage.PromptTokens += nr.AgentResult.Usage.PromptTokens
				totalUsage.CompletionTokens += nr.AgentResult.Usage.CompletionTokens
				totalUsage.TotalTokens += nr.AgentResult.Usage.TotalTokens
			}
			mu.Unlock()
		}

		// Determine next wave of ready nodes based on THIS wave's completions.
		mu.Lock()
		ready = g.findNextReady(ctx, waveCompleted, nodeResults, nodeOutputs, maxIter, iterations)
		mu.Unlock()
	}

	// Build output from terminal nodes.
	output := g.buildFinalOutput(nodeOutputs)

	runSpan.SetAttribute("graph.cost_usd", fmt.Sprintf("%.6f", totalCost))

	return &Result{
		RunID:       runID,
		Output:      output,
		NodeResults: nodeResults,
		TotalCost:   totalCost,
		TotalUsage:  totalUsage,
	}, nil
}

// runNode executes a single node's agent.
func (g *Graph) runNode(ctx context.Context, name, input string, iteration int) (*NodeResult, error) {
	_, nodeSpan := g.tracer.StartSpan(ctx, "graph.node")
	nodeSpan.SetAttribute("graph.node.name", name)
	nodeSpan.SetAttribute("graph.node.iteration", strconv.Itoa(iteration))
	defer g.tracer.EndSpan(nodeSpan)

	node := g.nodes[name]
	result, err := node.agent.Run(ctx, input)
	if err != nil {
		nodeSpan.SetError(err)
		return nil, fmt.Errorf("graph node %q: %w", name, err)
	}

	nodeSpan.SetAttribute("graph.node.cost_usd", fmt.Sprintf("%.6f", result.Cost))

	return &NodeResult{
		Name:        name,
		Output:      result.Message.Content,
		AgentResult: result,
		Iteration:   iteration,
	}, nil
}

// buildNodeInput determines the input for a node by merging outputs
// from all of its incoming predecessors.
func (g *Graph) buildNodeInput(name, initialInput string, outputs map[string]string) string {
	sources := g.incoming[name]
	if len(sources) == 0 {
		return initialInput
	}

	// Single incoming: use its output directly.
	if len(sources) == 1 {
		if out, ok := outputs[sources[0]]; ok {
			return out
		}
		return initialInput
	}

	// Multiple incoming (fan-in): merge all.
	var b strings.Builder
	for i, src := range sources {
		out, ok := outputs[src]
		if !ok {
			continue
		}
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString("[")
		b.WriteString(src)
		b.WriteString("]\n")
		b.WriteString(out)
	}
	return b.String()
}

// findNextReady evaluates edge conditions from nodes completed in the
// current wave and returns nodes that should run in the next wave.
func (g *Graph) findNextReady(
	ctx context.Context,
	waveCompleted map[string]bool,
	results map[string][]*NodeResult,
	outputs map[string]string,
	maxIter int,
	iterations map[string]int,
) []string {
	// Collect candidates from edges of nodes completed this wave.
	fired := make(map[string]map[string]bool) // candidate -> set of sources that fired

	for nodeName := range waveCompleted {
		for _, edge := range g.outgoing[nodeName] {
			// Evaluate condition.
			if edge.Condition != nil {
				nodeRes := g.latestResult(results, nodeName)
				if nodeRes == nil {
					continue
				}
				ok, err := edge.Condition(ctx, nodeRes)
				if err != nil || !ok {
					continue
				}
			}
			// Check iteration limit.
			if iterations[edge.To] >= maxIter {
				continue
			}
			if fired[edge.To] == nil {
				fired[edge.To] = make(map[string]bool)
			}
			fired[edge.To][nodeName] = true
		}
	}

	// For fan-in: a candidate is ready if all incoming sources that
	// completed in this wave and have unconditional edges have fired.
	// Sources that didn't run this wave don't block the candidate
	// (important for loops where only one branch re-fires).
	var nextReady []string
	for candidate, sources := range fired {
		ready := true
		for _, src := range g.incoming[candidate] {
			// Only check sources that completed this wave.
			if !waveCompleted[src] {
				continue
			}
			hasUnconditional := false
			for _, e := range g.outgoing[src] {
				if e.To == candidate && e.Condition == nil {
					hasUnconditional = true
					break
				}
			}
			if hasUnconditional && !sources[src] {
				ready = false
				break
			}
		}
		if ready {
			nextReady = append(nextReady, candidate)
		}
	}

	sort.Strings(nextReady)
	return nextReady
}

// latestResult returns the most recent result for a node.
func (g *Graph) latestResult(results map[string][]*NodeResult, name string) *NodeResult {
	rs := results[name]
	if len(rs) == 0 {
		return nil
	}
	return rs[len(rs)-1]
}

// buildFinalOutput joins outputs from terminal nodes.
func (g *Graph) buildFinalOutput(outputs map[string]string) string {
	if len(g.ends) == 1 {
		return outputs[g.ends[0]]
	}

	var parts []string
	for _, name := range g.ends {
		if out, ok := outputs[name]; ok {
			parts = append(parts, out)
		}
	}
	return strings.Join(parts, "\n\n")
}
