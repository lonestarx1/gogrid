// Package graph implements the Graph orchestration pattern.
//
// A graph extends the pipeline with conditional branches, parallel
// execution paths (fan-out/fan-in), and loops. Nodes wrap agents and
// edges connect them with optional condition functions that control
// routing. The execution engine performs topological traversal,
// running independent nodes concurrently and merging results at
// fan-in points.
//
// Graphs are built using a fluent builder API:
//
//	g, err := graph.NewBuilder("review").
//	    AddNode("draft", draftAgent).
//	    AddNode("review", reviewAgent).
//	    AddEdge("draft", "review").
//	    Build()
//
// Loops are supported with a max iteration guard to prevent infinite
// cycles. The graph can export to DOT format for Graphviz visualization.
package graph
